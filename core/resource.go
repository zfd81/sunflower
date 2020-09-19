package core

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/zfd81/rock/errs"

	"github.com/zfd81/rock/script/functions"

	"github.com/spf13/cast"

	"github.com/zfd81/rooster/util"

	"github.com/zfd81/rock/meta"

	"github.com/zfd81/rock/http"
	"github.com/zfd81/rock/script"
)

const (
	Regex     = "[A-Za-z0-9_.]+"
	LogFormat = "[LOG] %s "
)

type Context interface {
	GetDataSource(name string) script.DB
}

type ParrotResource struct {
	namespace     string              //命名空间 注:不能包含"/"
	context       Context             //上下文
	se            script.ScriptEngine // 脚本引擎
	method        string              // 资源请求方法
	path          string              // 资源原始路径
	regexPath     string              // 正则表达式形式路径
	level         int                 // 资源级别
	pathParams    []*meta.Parameter   // 路径参数
	requestParams []*meta.Parameter   // 请求参数
	log           *bytes.Buffer
	resp          *http.Response
}

func (r *ParrotResource) SetContext(context Context) {
	r.context = context
}

func (r *ParrotResource) GetMethod() string {
	return r.method
}

func (r *ParrotResource) GetPath() string {
	return r.path
}

func (r *ParrotResource) GetRegexPath() string {
	return r.regexPath
}

func (r *ParrotResource) GetLevel() int {
	return r.level
}

func (r *ParrotResource) GetPathParams() []*meta.Parameter {
	return r.pathParams
}

func (r *ParrotResource) AddPathParam(param *meta.Parameter) {
	r.pathParams = append(r.pathParams, param)
}

func (r *ParrotResource) GetRequestParams() []*meta.Parameter {
	return r.requestParams
}

func (r *ParrotResource) AddRequestParam(param *meta.Parameter) {
	r.requestParams = append(r.requestParams, param)
}

func (r *ParrotResource) GetNamespace() string {
	if r.namespace == "" {
		return meta.DefaultNamespace
	}
	return r.namespace
}

func (r *ParrotResource) SelectDataSource(name string) script.DB {
	return r.context.GetDataSource(name)
}

func (r *ParrotResource) Println(args ...interface{}) error {
	r.log.WriteString(fmt.Sprintf("[INFO] %s ", time.Now().Format("2006-01-02 15:04:05.000")))
	for _, arg := range args {
		r.log.WriteString(cast.ToString(arg))
	}
	r.log.WriteString("\n")
	return nil
}

func (r *ParrotResource) Perror(args ...interface{}) error {
	r.log.WriteString(fmt.Sprintf("[ERROR] %s ", time.Now().Format("2006-01-02 15:04:05.000")))
	for _, arg := range args {
		r.log.WriteString(errs.ErrorStyleFunc(cast.ToString(arg)))
	}
	r.log.WriteString("\n")
	return nil
}

func (r *ParrotResource) SetRespStatus(code int) {
	r.resp.SetStatusCode(code)
}

func (r *ParrotResource) AddRespHeader(name string, value interface{}) {
	r.resp.AddHeader(name, value)
}

func (r *ParrotResource) SetRespData(data interface{}) {
	r.resp.SetData(data)
}

func (r *ParrotResource) Run() (string, *http.Response, error) {
	for _, p := range r.pathParams {
		r.se.AddVar(p.Name, p.Value)
	}
	for _, p := range r.requestParams {
		r.se.AddVar(p.Name, p.Value)
	}
	err := r.se.Run()
	if err != nil {
		r.log.WriteString(fmt.Sprintf(LogFormat, time.Now().Format("2006-01-02 15:04:05.000")))
		r.log.WriteString(errs.ErrorStyleFunc(err))
		r.log.WriteString("\n")
	}
	return r.log.String(), r.resp, err
}

func (r *ParrotResource) Clear() {
	r.log.Reset()
	r.resp.Clear()
}

func NewResource(serv *meta.Service) *ParrotResource {
	path := serv.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}
	res := &ParrotResource{
		path:          path,
		pathParams:    []*meta.Parameter{},
		requestParams: []*meta.Parameter{},
	}
	regexPath, err := util.ReplaceBetween(path, "{", "}", func(i int, s int, e int, c string) (string, error) {
		param := &meta.Parameter{
			Name:     c,
			DataType: "string",
		}
		res.AddPathParam(param)
		return Regex, nil
	})
	if err != nil {
		return nil
	}
	se := script.New()
	se.SetScript(serv.Script)
	se.AddFunc("_sys_log", functions.SysLog(res))
	se.AddFunc("_sys_err", functions.SysError(res))
	se.AddFunc("_resp_write", functions.RespWrite(res))
	se.AddFunc("_http_get", functions.HttpGet)
	se.AddFunc("_http_post", functions.HttpPost)
	se.AddFunc("_http_delete", functions.HttpDelete)
	se.AddFunc("_http_put", functions.HttpPut)
	se.AddFunc("_db_query", functions.DBQuery(res))
	se.AddFunc("_db_queryOne", functions.DBQueryOne(res))
	se.AddFunc("_db_save", functions.DBSave(res))
	se.AddFunc("_db_exec", functions.DBExec(res))
	res.se = se
	res.method = strings.ToUpper(serv.Method)
	res.regexPath = regexPath
	pathFragments := strings.Split(regexPath, "/")
	res.level = len(pathFragments) - 1
	index := 0
	for i, fragment := range pathFragments {
		if Regex == fragment {
			res.pathParams[index].Index = i
			index++
		}
	}
	for _, p := range serv.Params {
		param := *p
		res.AddRequestParam(&param)
	}
	res.log = new(bytes.Buffer)
	res.resp = &http.Response{
		Header: map[string]string{},
	}
	return res
}