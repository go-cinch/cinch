package gen

import (
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen/biz"
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen/data"
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen/gorm"
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen/migrate"
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen/proto"
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen/service"
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen/sql"
	"github.com/spf13/cobra"
)

var CmdGen = &cobra.Command{
	Use:   "gen",
	Short: "gen: Generate Directory. gen gorm ",
	Long:  "gen: Generate Directory. gen gorm ",
}

func init() {
	CmdGen.AddCommand(gorm.CmdGorm)
	CmdGen.AddCommand(migrate.CmdMigrate)
	CmdGen.AddCommand(sql.CmdSql)
	CmdGen.AddCommand(proto.CmdProto)
	CmdGen.AddCommand(service.CmdService)
	CmdGen.AddCommand(biz.CmdBiz)
	CmdGen.AddCommand(data.CmdData)
}
