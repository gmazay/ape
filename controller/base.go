package controller

import (
	"ape/conf"
	"ape/db"

	"database/sql"
	//"github.com/kpango/glg"
	"ape/glg"
	//"net/http"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	//"strconv"
	"strings"
	//"regexp"
)

type Controller struct {
	cfg *conf.Config
	dbh *sql.DB
	log *glg.Glg
}

func NewController(cfg *conf.Config, dbh *sql.DB, log *glg.Glg) *Controller {
	c := new(Controller)
	c.cfg = cfg
	c.dbh = dbh
	c.log = log
	return c
}

func (c *Controller) Dispatcher(path string) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		method := string(ctx.Method())

		//fmt.Println(path, "--", method, "---\n")
		if c.cfg.Route[path] == nil || c.cfg.Route[path].Method[method] == nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			fmt.Fprintf(ctx, "Error in route configuration\n")
			c.log.Errorf("Error in route configuration: %s, %s %s", method, path, ctx.QueryArgs())
			return
		}

		//- Basic Auth ----//
		auth_res := make([]string, 2)
		if c.cfg.Route[path].Method[method].Auth == "basic" {
			auth_res = c.basicAuth(ctx.Request.Header.Peek("Authorization"))
			if len(auth_res) == 0 {
				fmt.Fprintf(ctx, "Authorization failed\n")
				ctx.SetStatusCode(fasthttp.StatusUnauthorized)
				ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Who\xA0are\xA0you?")
				c.log.Warnf("Authorization failed: %s %s %s %s", ctx.RemoteAddr(), method, path, ctx.QueryArgs())
				return
			}
		}
		//-----------------//

		switch c.cfg.Route[path].Method[method].Handler {
		case "FS":			c.FS(ctx, auth_res)
		case "GetGws":		c.GetGws(ctx, auth_res)
		case "Rates":		c.InFormOutArray(ctx, auth_res)
		case "ArrayInsert":	c.ArrayInsert(ctx, auth_res)
		case "PrintHTML":	c.PrintHTML(ctx, auth_res)

		default:			c.InFormOutArray(ctx, auth_res)
		}
		return
	})
}

func (c *Controller) FS(ctx *fasthttp.RequestCtx, auth_res []string) {
	path := string(ctx.Path())
	method := string(ctx.Method())

	//fmt.Println(c.cfg.DocumentRoot)

	//- Auth, check home -//
	if len(auth_res) < 2 || auth_res[1] == "" {
		fmt.Fprintf(ctx, "You not have a home! Go away!\n")
		c.log.Warnf("Wrong homedir: %s %s %s %s", ctx.RemoteAddr(), method, path, ctx.QueryArgs())
		return
	}
	//--------------------//
	uid := auth_res[0]
	dir := c.cfg.DocumentRoot + auth_res[1]

	fs := &fasthttp.FS{
		Root:               dir,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true,
		Compress:           false,
		AcceptByteRange:    false,
	}
	fs.PathRewrite = fasthttp.NewPathSlashesStripper(1)

	fsHandler := fs.NewRequestHandler()

	fsHandler(ctx)
	c.log.Logf("%s %s %s %s uid:%s", ctx.RemoteAddr(), method, path, ctx.QueryArgs(), uid)
}

func (c *Controller) InFormOutArray(ctx *fasthttp.RequestCtx, auth_res []string) {
	var query string
	var params []string
	var default_p []string
	var answer []string
	var parvals []interface{}

	path := string(ctx.Path())
	method := string(ctx.Method())

	if c.cfg.Route[path] != nil && c.cfg.Route[path].Method[method] != nil {
		query = c.cfg.Route[path].Method[method].Query
		params = c.cfg.Route[path].Method[method].Params
		default_p = c.cfg.Route[path].Method[method].Default

		answer = c.cfg.Route[path].Method[method].Answer
	} else {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		c.log.Errorf("Error in config route: %s, %s %s", method, path, ctx.QueryArgs())
		return
	}

	uid := "0"
	if len(auth_res) > 0 {
		uid = auth_res[0]
	}

	for i, param := range params {
		if len(ctx.FormValue(param)) == 0 {
			if len(default_p) > i && default_p[i] != "REQUIRED" {
				parvals = append(parvals, default_p[i])
			} else {
				ctx.SetStatusCode(fasthttp.StatusBadRequest)
				fmt.Fprintf(ctx, "\"error\":\"param '%s' is required!\"", param)
				c.log.Warnf("Bad request: no expected param: %s, %s %s %s", param, method, path, ctx.QueryArgs())
				return
			}
		} else {
			parvals = append(parvals, string(ctx.FormValue(param)))
		}
	}

	c.log.Logf("%s %s %s %s %s uid:%s", ctx.RemoteAddr(), method, path, ctx.QueryArgs(), parvals, uid)

	data, err := db.FetchAll2(c.dbh, query, parvals)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		c.log.Errorf("Error DB query: %s", err)
		fmt.Fprintf(ctx, "\"error\":\"Error DB query: %s\"", err)
		return
	}

	if string(ctx.FormValue("format")) == "csv" { // csv
		contentType := "text/plain; charset=UTF-8"
		if len(ctx.FormValue("download")) > 0 {
			contentType = "application/download"
		}
		ctx.SetContentType(contentType)
		ctx.SetStatusCode(fasthttp.StatusOK)

		fmt.Fprintf(ctx, "\"%s\"\n", strings.Join(answer, "\";\""))
		for _, row := range data {
			fmt.Fprintf(ctx, "\"%s\"\n", strings.Join(row, "\";\""))
		}

	} else { // json
		var res []map[string]string
		contentType := "application/json; charset=UTF-8"
		if len(ctx.FormValue("download")) > 0 {
			contentType = "application/download"
		}

		for _, r := range data {
			row := make(map[string]string)
			for j, s := range r {
				row[answer[j]] = s
			}
			res = append(res, row)
		}

		res_json, err := json.Marshal(res)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			c.log.Errorf("Error json encoding: %s %s %s %s %s", ctx.RemoteAddr(), method, path, ctx.QueryArgs(), parvals)
			return
		}
		//fmt.Fprintf( ctx, "%s", string(res_json) )
		ctx.SetContentType(contentType)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody([]byte(res_json))
	}

}

func (c *Controller) basicAuth(auth_header []byte) []string {
	//uid := 0
	if len(auth_header) == 0 {
		return []string{}
	}
	auth := strings.SplitN(string(auth_header), " ", 2)
	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		return []string{}
	}

	auth_res, err := db.FetchRow(c.dbh, c.cfg.Auth["query"], pair[0], pair[1])
	if err != nil {
		c.log.Errorf("Auth DB error: %s, %s:%s", err, pair[0], pair[1])
		return []string{}
	}

	return auth_res
}
