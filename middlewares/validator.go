package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/namelyzz/sayit/models"
	"reflect"
	"strings"
)

// 定义一个全局的翻译器
var trans ut.Translator

/*
InitTrans 表单验证，结合了不同语言的本地化支持
用于Gin框架中，结合结构体和自定义验证规则来验证用户提交的数据
locale 表示选择的语言，en 或 zh
下面是它的具体流程：
1. 通过 v.RegisterTagNameFunc 来自定义 JSON 标签的解析方法
2. 根据传入的 locale 参数选择适合的语言，然后获取相应语言的翻译器
3. 通过 RegisterStructValidation 注册结构体级别的验证规则
*/
func InitTrans(locale string) (err error) {
	// 获取 Gin 的验证器实例
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 自定义 JSON 标签提取函数，用于结构体字段的JSON标签
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// 提取结构体字段的 json 标签中的第一个部分
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" { // 表示忽略字段
				return ""
			}
			return name
		})

		// 为注册的结构体添加自定义的验证规则
		v.RegisterStructValidation(SignUpParamStructLevelValidation, models.ParamSignUp{})

		// 创建中英文翻译器实例
		zhT := zh.New()
		enT := en.New()

		// 创建支持中英文的多语言翻译器
		uni := ut.New(enT, zhT, enT)

		var ok bool
		// 获取指定语言的翻译器
		trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		/*
		   根据选择的语言注册默认的错误消息翻译。为什么添加翻译器？这一步主要是用于用户体验和国际化
		   例如：
		   如果用户提交表单时，Password 和 RePassword 不匹配，
		   Gin 会触发 SignUpParamStructLevelValidation 验证规则,报错：“Passwords do not match”。
		   如果用户的 locale 是 zh，翻译器会将该错误消息翻译成中文：“密码和确认密码不匹配”。
		*/
		switch locale {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, trans)
		default:
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		}
		return
	}
	return
}

// SignUpParamStructLevelValidation 自定义 SignUpParam 结构体校验函数
func SignUpParamStructLevelValidation(sl validator.StructLevel) {
	// 将当前验证的结构体转换为 ParamSignUp 类型
	su := sl.Current().Interface().(models.ParamSignUp)

	// 判断密码和确认密码是否相同
	if su.Password != su.RePassword {
		// 如果不相同，报告错误，错误的字段是 RePassword
		// 第二个参数是字段名，第三个是标签，第四个是规则，第五个是自定义的参数
		sl.ReportError(su.RePassword, "re_password", "RePassword", "eqfield", "password")
	}
}
