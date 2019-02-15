package controller

import (
	//"ape/conf"
	"ape/db"
	//"database/sql"
	//"ape/glg"
	//"net/http"
	//"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"strconv"
	//"strings"
	"regexp"
)

func (c *Controller) GetGws(ctx *fasthttp.RequestCtx, auth_res []string) {
	fmt.Println(ctx.QueryArgs())
	//fmt.Println("scheme: ", scheme)
	fmt.Println("path: ", string(ctx.Path()))

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)

	//	fmt.Println("--------------------------")
	//	for key, value := range ctx.QueryArgs() {
	//		fmt.Println(key, ":", value[0])
	//	}

	fmt.Println(c.cfg.Dsn["main"].User)

	id, err := strconv.Atoi(string(ctx.FormValue("id")))
	if err != nil {
		//id = 0
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		c.log.Warn("Bad request: no id")
		return
	}
	//id := string( ctx.FormValue("id") )
	//if id == "" { id = "1" }

	fmt.Println("id: ", id)

	///gws := make([][]string, 0)
	//gws, err = db.FetchAll( c.dbh, "select id, address, name, priority from gw where id=?", id )
	gws, err := db.FetchAll(c.dbh, "select plane_rate_id as id, plane_rate_name, plane_rate_info from plane_rate where plane_rate_id=$1", id)
	if err != nil {
		panic(err)
	}

	for _, r := range gws {
		//fmt.Println( r )
		for _, s := range r {
			fmt.Fprintf(ctx, "%s ", s)
		}
	}
}

func (c *Controller) ArrayInsert(ctx *fasthttp.RequestCtx, auth_res []string) {
	var query string
	var params []string
	//var default_p []string

	path := string(ctx.Path())
	method := string(ctx.Method())

	if c.cfg.Route[path] != nil && c.cfg.Route[path].Method[method] != nil {
		query = c.cfg.Route[path].Method[method].Query
		params = c.cfg.Route[path].Method[method].Params
		//default_p = c.cfg.Route[path].Method[method].Default
		//answer = c.cfg.Route[path].Method[method].Answer
	} else {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		c.log.Errorf("Error in config route: %s, %s %s", method, path, ctx.QueryArgs())
		return
	}

	uid := "0"
	if len(auth_res) > 0 {
		uid = auth_res[0]
	}

	//fmt.Fprintf(ctx, "--- %q ---\n", ctx.Path())
	//fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", ctx.Request.Body())

	var data []string

	switch method {
	case "PUT":
		if err := json.Unmarshal(ctx.Request.Body(), &data); err != nil {
			fmt.Fprintf(ctx, "Bad format: JSON array needed")
			c.log.Warnf("Bad request format: %s %s %s", method, path, ctx.Request.Body())
			return
		}
	case "POST":
		if len(params) == 0 || len(ctx.FormValue(params[0])) == 0 {
			fmt.Fprintf(ctx, "Not enougth data")
			c.log.Warnf("Not enougth data: %s %s uid:%s", method, path, uid)
			return
		} else {
			str := string(ctx.FormValue(params[0]))
			data = regexp.MustCompile(`\s*\r*\n\s*`).Split(str, -1)
			//data = strings.Split( string(str), "\n" )
		}
	}
	fmt.Println(data)

	c.log.Logf("%s %s %s %s %s uid:%s", ctx.RemoteAddr(), method, path, ctx.QueryArgs(), ctx.Request.Body(), uid)

	var validPrefix = regexp.MustCompile(`^\d+$`)
	i := 0
	for _, number := range data {
		if !validPrefix.MatchString(number) {
			c.log.Errorf("Invalid format: %q", number)
			fmt.Fprintf(ctx, "Invalid format: %q\n", number)
			continue
		}
		err := db.Do(c.dbh, query, number)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			c.log.Errorf("Error DB query: %s", err)
			fmt.Fprintf(ctx, "Error DB query: \"%s\"\n", err)
			//return
		} else {
			i++
		}
	}

	fmt.Fprintf(ctx, "Inserted: %d of %d\n", i, len(data))
	ctx.SetContentType("text/plain; charset=UTF-8")
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (c *Controller) PrintHTML(ctx *fasthttp.RequestCtx, auth_res []string) {
	path := string(ctx.Path())
	method := string(ctx.Method())
	uid := "0"
	if len(auth_res) > 0 {
		uid = auth_res[0]
	}

	ctx.SetContentType("text/html; charset=UTF-8")
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "<html><head><meta http-equiv=Content-Type content=\"text/html; charset=UTF-8\"></head><body>")
	fmt.Fprintf(ctx, c.cfg.Route[path].Method[method].Query)
	fmt.Fprintf(ctx, "</body>")

	c.log.Logf("%s %s %s %s uid:%s", ctx.RemoteAddr(), method, path, ctx.QueryArgs(), uid)
}
