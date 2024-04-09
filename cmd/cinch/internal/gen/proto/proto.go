package proto

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
	DefaultSuffix = "proto"
	DefaultCover  = false
)

var CmdProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate proto file. Example: cinch gen proto -p api/game-proto/game.proto",
	Long:  "Generate proto file, contains basic CRUD api. Example: cinch gen proto -p api/game-proto/game.proto",
	Run:   run,
}

func init() {
	CmdProto.PersistentFlags().StringP("path", "p", DefaultPath, "generate file path")
	CmdProto.PersistentFlags().StringP("module", "m", DefaultModule, "module name")
	CmdProto.PersistentFlags().StringP("api", "a", DefaultApi, "api name(default same as module)")
	CmdProto.PersistentFlags().StringP("suffix", "s", DefaultSuffix, "generate dir suffix")
	CmdProto.PersistentFlags().BoolP("cover", "c", DefaultCover, "cover old file or not")
}

func run(cmd *cobra.Command, _ []string) {
	dir, _ := cmd.Flags().GetString("path")
	module, _ := cmd.Flags().GetString("module")
	api, _ := cmd.Flags().GetString("api")
	suffix, _ := cmd.Flags().GetString("suffix")
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
		dir = fmt.Sprintf("api/%s-%s/%s.proto", module, suffix, module)
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
			fmt.Fprintf(os.Stderr, "\033[31mERROR: file %s exist, pls change name or set cover=true, Example: cinch gen proto -c\033[m\n", dir)
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

	content := fmt.Sprintf(`syntax = "proto3";

package %v.v1;

import "openapiv3/annotations.proto";
import "google/api/annotations.proto";
import "google/api/field_behavior.proto";
import "google/protobuf/empty.proto";
import "cinch/params/params.proto";

option go_package = "api/%v;%v";
option java_multiple_files = true;
option java_package = "%v.v1";
option java_outer_classname = "%vProtoV1";

option (openapi.v3.document) = {
  info: {
    title: "%s service";
    version: "1.0.0";
    description: "This is %s service docs";
  }
  components: {
    security_schemes: {
      additional_properties: [
        {
          name: "BearerAuth";
          value: {
            security_scheme: {
              type: "http";
              scheme: "bearer";
            }
          }
        }
      ]
    }
  }
  security: [
    {
      additional_properties: [
        {
          name: "BearerAuth";
          value: {
            value: []
          }
        }
      ]
    }
  ]
};

service %v {
  // create one %v record
  rpc Create%v (Create%vRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      post: "/%v/create"
      body: "*"
    };
  }
  // query one %v record
  rpc Get%v (Get%vRequest) returns (Get%vReply) {
    option (google.api.http) = {
      get: "/%v/get/{id}"
    };
  }
  // query %v list by page
  rpc Find%v (Find%vRequest) returns (Find%vReply) {
    option (google.api.http) = {
      get: "/%v/list"
    };
  }
  // update one %v record by id
  rpc Update%v (Update%vRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      put: "/%v/update/{id}"
      body: "*",
      additional_bindings {
        patch: "/%v/update/{id}",
        body: "*",
      }
    };
  }
  // delete one or more %v record by id
  rpc Delete%v (params.IdsRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/%v/delete"
    };
  }
}

message %vReply {
  uint64 id = 1;
  string name = 2;
}

message Create%vRequest {
  string name = 1 [(google.api.field_behavior) = REQUIRED];
}

message Get%vRequest {
  uint64 id = 1 [(google.api.field_behavior) = REQUIRED];
}

message Get%vReply {
  uint64 id = 1;
  string name = 2;
}

message Find%vRequest {
  params.Page page = 1;
  optional string name = 2;
}

message Find%vReply {
  params.Page page = 1;
  repeated %vReply list = 2;
}

message Update%vRequest {
  uint64 id = 1 [(google.api.field_behavior) = REQUIRED];
  optional string name = 2;
}
`,
		// import
		module,
		module, module,
		module,
		camelModule,

		// openapi.v3
		camelModule,
		module,

		// service
		camelModule,

		// create
		camelApi,
		camelApi, camelApi,
		api,

		// get
		camelApi,
		camelApi, camelApi, camelApi,
		api,

		// find
		camelApi,
		camelApi, camelApi, camelApi,
		api,

		// update
		camelApi,
		camelApi, camelApi,
		api,
		api,

		// delete
		camelApi,
		camelApi,
		api,

		// message
		camelApi,
		camelApi,
		camelApi,
		camelApi,
		camelApi,

		camelApi, camelApi,

		camelApi,
	)

	_, err = f.Write([]byte(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: cannot insert file content: %s\033[m\n", err.Error())
		return
	}
	fmt.Printf("\n🍺 Generate proto file success: %s\n", color.GreenString(dir))
}
