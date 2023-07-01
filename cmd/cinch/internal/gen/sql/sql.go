package sql

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const (
	DefaultPath   = "internal/db/migrations"
	DefaultName   = "migration"
	DefaultTable  = "my_table"
	DefaultLayout = "2006010215"
	DefaultCover  = false
)

var CmdSql = &cobra.Command{
	Use:   "sql",
	Short: "Generate sql migration file by current timestamp. Example: cinch gen sql -n game -t game",
	Long:  "Generate sql migration file by current timestamp. Example: cinch gen sql -n game -t game",
	Run:   run,
}

func init() {
	CmdSql.PersistentFlags().StringP("path", "p", DefaultPath, "generate file path")
	CmdSql.PersistentFlags().StringP("name", "n", DefaultName, "generate filename suffix")
	CmdSql.PersistentFlags().StringP("table", "t", DefaultTable, "generate sql content table name")
	CmdSql.PersistentFlags().StringP("layout", "l", DefaultLayout, "generate filename timestamp layout")
	CmdSql.PersistentFlags().BoolP("cover", "c", DefaultCover, "cover old file or not")
}

func run(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("path")
	name, _ := cmd.Flags().GetString("name")
	table, _ := cmd.Flags().GetString("table")
	layout, _ := cmd.Flags().GetString("layout")
	cover, _ := cmd.Flags().GetBool("cover")
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: invalid path %s: %s\033[m\n", dir, err.Error())
		return
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: path %s is not a dir\033[m\n", dir)
		return
	}

	now := carbon.Now().Layout(layout)
	// .Format("2006010215")
	filename := strings.Join([]string{dir, "/", now, "-", name, ".sql"}, "")

	if !cover {
		_, err = os.Stat(filename)
		if err == nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: file %s exist, pls change name or set cover=true, Example: cinch gen sql -n game || cinch gen sql -c\033[m\n", filename)
			return
		}
	}

	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot create file %s, pls check permission: %s\033[m\n", filename, err.Error())
		return
	}

	content := strings.Join([]string{
		"-- +migrate Up",
		"CREATE TABLE `" + table + "`",
		"(",
		"  `id`         BIGINT UNSIGNED AUTO_INCREMENT COMMENT 'auto increment id' PRIMARY KEY,",
		"  `created_at` DATETIME(3) NULL COMMENT 'create time',",
		"  `updated_at` DATETIME(3) NULL COMMENT 'update time',",
		"  -- enable soft delete, do this:",
		"  -- `deleted_at` DATETIME(3) NULL COMMENT 'delete time',",
		"  `name` VARCHAR(50) COMMENT 'name'",
		"  -- `other_field` VARCHAR(50) COMMENT 'your field comment'",
		") ENGINE = InnoDB",
		"  DEFAULT CHARSET = utf8mb4",
		"  COLLATE = utf8mb4_general_ci;",
		"",
		"-- create table index, do this:",
		"-- CREATE UNIQUE INDEX `idx_name` ON `" + table + "` (`name`);",
		"",
		"-- +migrate Down",
		"DROP TABLE `" + table + "`;",
		"",
	}, "\n")

	_, err = f.Write([]byte(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot insert file content: %s\033[m\n", err.Error())
		return
	}
	fmt.Printf("\n🍺 Generate sql migration file success: %s\n", color.GreenString(filename))
}
