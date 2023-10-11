package generator

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util/console"
	"github.com/zeromicro/go-zero/tools/goctl/util/ctx"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

type ZRpcContext struct {
	// Sre is the source file of the proto.
	Src string
	// ProtoCmd is the command to generate proto files.
	ProtocCmd string
	// ProtoGenGrpcDir is the directory to store the generated proto files.
	ProtoGenGrpcDir string
	// ProtoGenGoDir is the directory to store the generated go files.
	ProtoGenGoDir string
	// IsGooglePlugin is the flag to indicate whether the proto file is generated by google plugin.
	IsGooglePlugin bool
	// GoOutput is the output directory of the generated go files.
	GoOutput string
	// GrpcOutput is the output directory of the generated grpc files.
	GrpcOutput string
	// Output is the output directory of the generated files.
	Output string
	// Multiple is the flag to indicate whether the proto file is generated in multiple mode.
	Multiple bool
}

// Generate generates a rpc service, through the proto file,
// code storage directory, and proto import parameters to control
// the source file and target location of the rpc service that needs to be generated
func (g *Generator) Generate(zctx *ZRpcContext) error {
	abs, err := filepath.Abs(zctx.Output)
	if err != nil {
		return err
	}
	startTime := time.Now()

	err = pathx.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	fmt.Println("MkdirIfNotExist，耗时:", time.Since(startTime))

	err = g.Prepare()
	if err != nil {
		return err
	}

	fmt.Println("g.Prepare,耗时", time.Since(startTime))

	projectCtx, err := ctx.Prepare(abs)
	if err != nil {
		return err
	}

	fmt.Println("Prepare，耗时", time.Since(startTime))

	p := parser.NewDefaultProtoParser()
	proto, err := p.Parse(zctx.Src, zctx.Multiple)
	if err != nil {
		return err
	}

	fmt.Println("p.Parse，耗时", time.Since(startTime))

	dirCtx, err := mkdir(projectCtx, proto, g.cfg, zctx)
	if err != nil {
		return err
	}

	fmt.Println("mkdir，耗时", time.Since(startTime))

	err = g.GenEtc(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	fmt.Println("GenEtc，耗时", time.Since(startTime))

	err = g.GenPb(dirCtx, zctx)
	if err != nil {
		return err
	}

	fmt.Println("GenPb，耗时", time.Since(startTime))

	err = g.GenConfig(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	fmt.Println("GenConfig，耗时", time.Since(startTime))

	err = g.GenSvc(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	fmt.Println("GenSvc，耗时", time.Since(startTime))

	err = g.GenLogic(dirCtx, proto, g.cfg, zctx)
	if err != nil {
		return err
	}

	fmt.Println("GenLogic，耗时", time.Since(startTime))

	err = g.GenServer(dirCtx, proto, g.cfg, zctx)
	if err != nil {
		return err
	}

	fmt.Println("GenServer，耗时", time.Since(startTime))

	err = g.GenMain(dirCtx, proto, g.cfg, zctx)
	if err != nil {
		return err
	}

	fmt.Println("GenMain，耗时", time.Since(startTime))

	err = g.GenCall(dirCtx, proto, g.cfg, zctx)

	fmt.Println("GenCall，耗时", time.Since(startTime))

	//rpc
	cmd := exec.Command("goimports", "-w", zctx.Output)
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Printf("goimports -w %s，耗时: %v\n", zctx.Output, time.Since(startTime))

	//proto
	protoPath := filepath.Dir(zctx.Src)
	cmd = exec.Command("goimports", "-w", protoPath)
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Printf("goimports -w %s，耗时: %v\n", protoPath, time.Since(startTime))

	console.NewColorConsole().MarkDone()

	fmt.Println("总耗时", time.Since(startTime))

	return err
}
