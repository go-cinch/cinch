package service

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

var CmdService = &cobra.Command{
	Use:   "service",
	Short: "Generate service file. Example: cinch gen service -p internal/service/game.go",
	Long:  "Generate service file, contains basic CRUD api. Example: cinch gen service -p internal/service/game.go",
	Run:   run,
}

func init() {
	CmdService.PersistentFlags().StringP("path", "p", DefaultPath, "generate file path")
	CmdService.PersistentFlags().StringP("module", "m", DefaultModule, "module name")
	CmdService.PersistentFlags().StringP("api", "a", DefaultApi, "api name(default same as module)")
	CmdService.PersistentFlags().BoolP("cover", "c", DefaultCover, "cover old file or not")
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
		dir = fmt.Sprintf("internal/service/%s.go", api)
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
			fmt.Fprintf(os.Stderr, "\033[31mERROR: file %s exist, pls change name or set cover=true, Example: cinch gen service -c\033[m\n", dir)
			return
		}
	}

	f, err := os.Create(dir)
	defer f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot create file %s, pls check permission: %s\033[m\n", dir, err.Error())
		return
	}

	camelModule := utils.CamelCase(module)
	camelApi := utils.CamelCase(api)

	content := fmt.Sprintf(`package service

import (
	"context"

	"github.com/go-cinch/common/copierx"
	"github.com/go-cinch/common/page"
	"github.com/go-cinch/common/proto/params"
	"github.com/go-cinch/common/utils"
	"%v/api/%v"
	"%v/internal/biz"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *%vService) Create%v(ctx context.Context, req *%v.Create%vRequest) (rp *emptypb.Empty, err error) {
	tr := otel.Tracer("api")
	ctx, span := tr.Start(ctx, "Create%v")
	defer span.End()
	rp = &emptypb.Empty{}
	r := &biz.%v{}
	copierx.Copy(&r, req)
	err = s.%v.Create(ctx, r)
	return
}

func (s *%vService) Get%v(ctx context.Context, req *%v.Get%vRequest) (rp *%v.Get%vReply, err error) {
	tr := otel.Tracer("api")
	ctx, span := tr.Start(ctx, "Get%v")
	defer span.End()
	rp = &%v.Get%vReply{}
	res, err := s.%v.Get(ctx, req.Id)
	if err != nil {
		return
	}
	copierx.Copy(&rp, res)
	return
}

func (s *%vService) Find%v(ctx context.Context, req *%v.Find%vRequest) (rp *%v.Find%vReply, err error) {
	tr := otel.Tracer("api")
	ctx, span := tr.Start(ctx, "Find%v")
	defer span.End()
	rp = &%v.Find%vReply{}
	rp.Page = &params.Page{}
	r := &biz.Find%v{}
	r.Page = page.Page{}
	copierx.Copy(&r, req)
	copierx.Copy(&r.Page, req.Page)
	res := s.%v.Find(ctx, r)
	copierx.Copy(&rp.Page, r.Page)
	copierx.Copy(&rp.List, res)
	return
}

func (s *%vService) Update%v(ctx context.Context, req *%v.Update%vRequest) (rp *emptypb.Empty, err error) {
	tr := otel.Tracer("api")
	ctx, span := tr.Start(ctx, "Update%v")
	defer span.End()
	rp = &emptypb.Empty{}
	r := &biz.Update%v{}
	copierx.Copy(&r, req)
	err = s.%v.Update(ctx, r)
	return
}

func (s *%vService) Delete%v(ctx context.Context, req *params.IdsRequest) (rp *emptypb.Empty, err error) {
	tr := otel.Tracer("api")
	ctx, span := tr.Start(ctx, "Delete%v")
	defer span.End()
	rp = &emptypb.Empty{}
	err = s.%v.Delete(ctx, utils.Str2Uint64Arr(req.Ids)...)
	return
}
`,
		module, module, module, camelModule, camelApi,
		module, camelApi, camelModule, camelApi, api,

		camelModule, camelApi, module, camelApi, module,
		camelApi, camelApi, module, camelApi, api,

		camelModule, camelApi, module, camelApi, module,
		camelApi, camelApi, module, camelApi, camelApi,

		api, camelModule, camelApi, module, camelApi,
		camelApi, camelApi, api, camelModule, camelApi,

		camelApi, api,
	)

	_, err = f.Write([]byte(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot insert file content: %s\033[m\n", err.Error())
		return
	}

	base.Lint(fileDir)

	fmt.Printf("\n🍺 Generate service file success: %s\n", color.GreenString(dir))
}
