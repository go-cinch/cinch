package biz

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-cinch/cinch/cmd/cinch/internal/base"
	"github.com/go-cinch/common/utils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	DefaultPath   = ""
	DefaultModule = ""
	DefaultApi    = ""
	DefaultCover  = false
)

var CmdBiz = &cobra.Command{
	Use:   "biz",
	Short: "Generate biz file. Example: cinch gen biz -p internal/biz/game.go",
	Long:  "Generate biz file, contains basic CRUD api. Example: cinch gen biz -p internal/biz/game.go",
	Run:   run,
}

func init() {
	CmdBiz.PersistentFlags().StringP("path", "p", DefaultPath, "generate file path")
	CmdBiz.PersistentFlags().StringP("module", "m", DefaultModule, "module name")
	CmdBiz.PersistentFlags().StringP("api", "a", DefaultApi, "api name(default same as module)")
	CmdBiz.PersistentFlags().BoolP("cover", "c", DefaultCover, "cover old file or not")
}

func run(cmd *cobra.Command, _ []string) {
	dir, _ := cmd.Flags().GetString("path")
	module, _ := cmd.Flags().GetString("module")
	api, _ := cmd.Flags().GetString("api")
	cover, _ := cmd.Flags().GetBool("cover")

	var err error
	if module == DefaultModule {
		module, err = base.ModulePath("go.mod")
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot find go.mod: %s\033[m\n", err.Error())
			return
		}
	}
	if api == DefaultApi {
		api = module
	}
	if dir == DefaultPath {
		dir = fmt.Sprintf("internal/biz/%s.go", api)
	}

	fileDir, _ := filepath.Split(dir)

	err = os.MkdirAll(fileDir, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot create dir %s:%s\033[m\n", fileDir, err.Error())
		return
	}

	if !cover {
		_, err = os.Stat(dir)
		if err == nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: file %s exist, pls change name or set cover=true, Example: cinch gen biz -c\033[m\n", dir)
			return
		}
	}

	f, err := os.Create(dir)
	defer f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot create file %s, pls check permission: %s\033[m\n", dir, err.Error())
		return
	}

	camelApi := utils.CamelCase(api)

	content := fmt.Sprintf(`package biz

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-cinch/common/constant"
	"github.com/go-cinch/common/copierx"
	"github.com/go-cinch/common/middleware/i18n"
	"github.com/go-cinch/common/page"
	"github.com/go-cinch/common/utils"
	"%v/api/reason"
	"%v/internal/conf"
	"github.com/pkg/errors"
)

type %v struct {
	Id   uint64 %vjson:"id,string"%v
	Name string %vjson:"name"%v
}

type Find%v struct {
	Page page.Page %vjson:"page"%v
	Name *string   %vjson:"name"%v
}

type Find%vCache struct {
	Page page.Page %vjson:"page"%v
	List []%v    %vjson:"list"%v
}

type Update%v struct {
	Id   uint64  %vjson:"id,string"%v
	Name *string %vjson:"name,omitempty"%v
}

type %vRepo interface {
	Create(ctx context.Context, item *%v) error
	Get(ctx context.Context, id uint64) (*%v, error)
	Find(ctx context.Context, condition *Find%v) []%v
	Update(ctx context.Context, item *Update%v) error
	Delete(ctx context.Context, ids ...uint64) error
}

type %vUseCase struct {
	c     *conf.Bootstrap
	repo  %vRepo
	tx    Transaction
	cache Cache
}

func New%vUseCase(c *conf.Bootstrap, repo %vRepo, tx Transaction, cache Cache) *%vUseCase {
	return &%vUseCase{
		c:    c,
		repo: repo,
		tx:   tx,
		cache: cache.WithPrefix(strings.Join([]string{
			c.Name, "%v",
		}, "_")),
	}
}

func (uc *%vUseCase) Create(ctx context.Context, item *%v) error {
	return uc.tx.Tx(ctx, func(ctx context.Context) error {
		return uc.cache.Flush(ctx, func(ctx context.Context) error {
			return uc.repo.Create(ctx, item)
		})
	})
}

func (uc *%vUseCase) Get(ctx context.Context, id uint64) (rp *%v, err error) {
	rp = &%v{}
	action := strings.Join([]string{"get", strconv.FormatUint(id, 10)}, "_")
	str, ok := uc.cache.Get(ctx, action, func(ctx context.Context) (string, bool) {
		return uc.get(ctx, action, id)
	})
	if ok {
		utils.Json2Struct(&rp, str)
		if rp.Id == constant.UI0 {
			err = reason.ErrorNotFound("%%s %v.id: %%d", i18n.FromContext(ctx).T(RecordNotFound), id)
		}
		return
	}
	err = reason.ErrorTooManyRequests(i18n.FromContext(ctx).T(TooManyRequests))
	return
}

func (uc *%vUseCase) get(ctx context.Context, action string, id uint64) (res string, ok bool) {
	// read data from db and write to cache
	rp := &%v{}
	item, err := uc.repo.Get(ctx, id)
	notFound := errors.Is(err, reason.ErrorNotFound(i18n.FromContext(ctx).T(RecordNotFound)))
	if err != nil && !notFound {
		return
	}
	copierx.Copy(&rp, item)
	res = utils.Struct2Json(rp)
	uc.cache.Set(ctx, action, res, notFound)
	ok = true
	return
}

func (uc *%vUseCase) Find(ctx context.Context, condition *Find%v) (rp []%v) {
	// use md5 string as cache replay json str, key is short
	action := strings.Join([]string{"find", utils.StructMd5(condition)}, "_")
	str, ok := uc.cache.Get(ctx, action, func(ctx context.Context) (string, bool) {
		return uc.find(ctx, action, condition)
	})
	if ok {
		var cache Find%vCache
		utils.Json2Struct(&cache, str)
		condition.Page = cache.Page
		rp = cache.List
	}
	return
}

func (uc *%vUseCase) find(ctx context.Context, action string, condition *Find%v) (res string, ok bool) {
	// read data from db and write to cache
	list := uc.repo.Find(ctx, condition)
	var cache Find%vCache
	cache.List = list
	cache.Page = condition.Page
	res = utils.Struct2Json(cache)
	uc.cache.Set(ctx, action, res, len(list) == 0)
	ok = true
	return
}

func (uc *%vUseCase) Update(ctx context.Context, item *Update%v) error {
	return uc.tx.Tx(ctx, func(ctx context.Context) error {
		return uc.cache.Flush(ctx, func(ctx context.Context) (err error) {
			err = uc.repo.Update(ctx, item)
			return
		})
	})
}

func (uc *%vUseCase) Delete(ctx context.Context, ids ...uint64) error {
	return uc.tx.Tx(ctx, func(ctx context.Context) error {
		return uc.cache.Flush(ctx, func(ctx context.Context) (err error) {
			err = uc.repo.Delete(ctx, ids...)
			return
		})
	})
}
`,
		module, module, camelApi, "`", "`",
		"`", "`", camelApi, "`", "`",

		"`", "`", camelApi, "`", "`",
		camelApi, "`", "`", camelApi, "`",

		"`", "`", "`", camelApi, camelApi,
		camelApi, camelApi, camelApi, camelApi, camelApi,

		camelApi, camelApi, camelApi, camelApi, camelApi,
		api, camelApi, camelApi, camelApi, camelApi,

		camelApi, camelApi, camelApi, camelApi, camelApi,
		camelApi, camelApi, camelApi, camelApi, camelApi,

		camelApi, camelApi, camelApi, camelApi,
	)

	_, err = f.Write([]byte(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot insert file content: %s\033[m\n", err.Error())
		return
	}

	base.Lint(fileDir)

	fmt.Printf("\n🍺 Generate biz file success: %s\n", color.GreenString(dir))
}
