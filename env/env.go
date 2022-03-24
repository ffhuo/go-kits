package env

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	active Environment
	dev    Environment = &environment{value: "dev"}
	fat    Environment = &environment{value: "fat"}
	uat    Environment = &environment{value: "uat"}
	pro    Environment = &environment{value: "pro"}
)

var _ Environment = (*environment)(nil)

// Environment 环境配置
type Environment interface {
	Value() string
	File() string
	SetFile(file string)
	IsDev() bool
	IsFat() bool
	IsUat() bool
	IsPro() bool
	t()
}

type environment struct {
	value string
	file  string
}

func (e *environment) Value() string {
	return e.value
}

func (e *environment) File() string {
	return e.file
}

func (e *environment) SetFile(file string) {
	e.file = file
}

func (e *environment) IsDev() bool {
	return e.value == "dev"
}

func (e *environment) IsFat() bool {
	return e.value == "fat"
}

func (e *environment) IsUat() bool {
	return e.value == "uat"
}

func (e *environment) IsPro() bool {
	return e.value == "pro"
}

func (e *environment) t() {}

func init() {
	env := flag.String("env", "", "请输入运行环境:\n dev:开发环境\n fat:测试环境\n uat:预上线环境\n pro:正式环境\n")
	file := flag.String("config", "", "请输入运行配置文件\n")
	if strings.HasSuffix(os.Args[0], ".test") || (len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-test")) {

	} else {
		flag.Parse()
	}

	switch strings.ToLower(strings.TrimSpace(*env)) {
	case "dev":
		active = dev
	case "fat":
		active = fat
	case "uat":
		active = uat
	case "pro":
		active = pro
	default:
		active = dev
		fmt.Println("Warning: '-env' cannot be found, or it is illegal. The default 'dev' will be used.")
	}
	if *file != "" {
		active.SetFile(*file)
	}
}

// Active 当前配置的env
func Active() Environment {
	return active
}
