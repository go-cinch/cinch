package data

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

var CmdData = &cobra.Command{
	Use:   "data",
	Short: "Generate data file. Example: cinch gen data -p internal/data/game.go",
	Long:  "Generate data file, contains basic CRUD api. Example: cinch gen data -p internal/data/game.go",
	Run:   run,
}

func init() {
	CmdData.PersistentFlags().StringP("path", "p", DefaultPath, "generate file path")
	CmdData.PersistentFlags().StringP("module", "m", DefaultModule, "module name")
	CmdData.PersistentFlags().StringP("api", "a", DefaultApi, "api name(default same as module)")
	CmdData.PersistentFlags().BoolP("cover", "c", DefaultCover, "cover old file or not")
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
		dir = fmt.Sprintf("internal/data/%s.go", api)
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
			fmt.Fprintf(os.Stderr, "\033[31mERROR: file %s exist, pls change name or set cover=true, Example: cinch gen data -c\033[m\n", dir)
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

	content := fmt.Sprintf(`package data

import (
	"context"
	"strings"

	"github.com/go-cinch/common/constant"
	"github.com/go-cinch/common/copierx"
	"github.com/go-cinch/common/log"
	"github.com/go-cinch/common/utils"
	"%v/internal/biz"
	"%v/internal/data/model"
	"%v/internal/data/query"
	"gorm.io/gen"
)

type %vRepo struct {
	data *Data
}

func New%vRepo(data *Data) biz.%vRepo {
	return &%vRepo{
		data: data,
	}
}

func (ro %vRepo) Create(ctx context.Context, item *biz.%v) (err error) {
	err = ro.NameExists(ctx, item.Name)
	if err == nil {
		err = biz.ErrDuplicateField(ctx, "name", item.Name)
		return
	}
	var m model.%v
	copierx.Copy(&m, item)
	p := query.Use(ro.data.DB(ctx)).%v
	db := p.WithContext(ctx)
	m.ID = ro.data.Id(ctx)
	err = db.Create(&m)
	return
}

func (ro %vRepo) Get(ctx context.Context, id uint64) (item *biz.%v, err error) {
	item = &biz.%v{}
	p := query.Use(ro.data.DB(ctx)).%v
	db := p.WithContext(ctx)
	m := db.GetByID(id)
	if m.ID == constant.UI0 {
		err = biz.ErrRecordNotFound(ctx)
		return
	}
	copierx.Copy(&item, m)
	return
}

func (ro %vRepo) Find(ctx context.Context, condition *biz.Find%v) (rp []biz.%v) {
	p := query.Use(ro.data.DB(ctx)).%v
	db := p.WithContext(ctx)
	rp = make([]biz.%v, 0)
	list := make([]model.%v, 0)
	conditions := make([]gen.Condition, 0, 2)
	if condition.Name != nil {
		conditions = append(conditions, p.Name.Like(strings.Join([]string{"%%", *condition.Name, "%%"}, "")))
	}
	condition.Page.Primary = "id"
	condition.Page.
		WithContext(ctx).
		Query(
			db.
				Order(p.ID.Desc()).
				Where(conditions...).
				UnderlyingDB(),
		).
		Find(&list)
	copierx.Copy(&rp, list)
	return
}

func (ro %vRepo) Update(ctx context.Context, item *biz.Update%v) (err error) {
	p := query.Use(ro.data.DB(ctx)).%v
	db := p.WithContext(ctx)
	m := db.GetByID(item.Id)
	if m.ID == constant.UI0 {
		err = biz.ErrRecordNotFound(ctx)
		return
	}
	change := make(map[string]interface{})
	utils.CompareDiff(m, item, &change)
	if len(change) == 0 {
		err = biz.ErrDataNotChange(ctx)
		return
	}
	if item.Name != nil && *item.Name != m.Name {
		err = ro.NameExists(ctx, *item.Name)
		if err == nil {
			err = biz.ErrDuplicateField(ctx, "name", *item.Name)
			return
		}
	}
	_, err = db.
		Where(p.ID.Eq(item.Id)).
		Updates(&change)
	return
}

func (ro %vRepo) Delete(ctx context.Context, ids ...uint64) (err error) {
	p := query.Use(ro.data.DB(ctx)).%v
	db := p.WithContext(ctx)
	_, err = db.
		Where(p.ID.In(ids...)).
		Delete()
	return
}

func (ro %vRepo) NameExists(ctx context.Context, name string) (err error) {
	p := query.Use(ro.data.DB(ctx)).%v
	db := p.WithContext(ctx)
	arr := strings.Split(name, ",")
	for _, item := range arr {
		res := db.GetByCol("name", item)
		if res.ID == constant.UI0 {
			err = biz.ErrRecordNotFound(ctx)
			log.
				WithError(err).
				Error("invalid %vname%v: %%s", name)
			return
		}
	}
	return
}
`,
		module, module, module, api, camelApi,
		camelApi, api, api, camelApi, camelApi,

		camelApi, api, camelApi, camelApi, camelApi,
		api, camelApi, camelApi, camelApi, camelApi,

		camelApi, api, camelApi, camelApi, api,
		camelApi, api, camelApi, "`", "`",
	)

	_, err = f.Write([]byte(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot insert file content: %s\033[m\n", err.Error())
		return
	}

	base.Lint(fileDir)

	fmt.Printf("\n🍺 Generate data file success: %s\n", color.GreenString(dir))
}
