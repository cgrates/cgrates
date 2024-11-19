//go: build ignore

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package main

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"unicode"

	"github.com/cgrates/cgrates/accounts"
	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

func main() {
	type genFile struct {
		path       string
		subsystem  string // the name of the constant
		obj        any
		customName string
	}
	fmt.Println("Generating dispatcher files ...")
	for _, file := range []genFile{
		{"accounts.go", "MetaAccounts", new(accounts.AccountS), utils.EmptyString},
		{"actions.go", "MetaActions", new(actions.ActionS), utils.EmptyString},
		{"attributes.go", "MetaAttributes", new(engine.AttributeS), utils.EmptyString},
		{"caches.go", "MetaCaches", engine.Cache, utils.EmptyString},
		{"cdrs.go", "MetaCDRs", new(cdrs.CDRServer), utils.CDRs},
		{"chargers.go", "MetaChargers", new(engine.ChargerS), utils.EmptyString},
		{"config.go", "MetaConfig", new(config.CGRConfig), utils.ConfigS},
		{"rates.go", "RateS", new(rates.RateS), utils.EmptyString},
		{"replicator.go", "MetaReplicator", new(apis.ReplicatorSv1), utils.EmptyString},
		{"resources.go", "MetaResources", new(engine.ResourceS), utils.EmptyString},
		{"routes.go", "MetaRoutes", new(engine.RouteS), utils.EmptyString},
		{"sessions.go", "MetaSessionS", new(sessions.SessionS), utils.SessionS},
		{"rankings.go", "MetaRankings", new(engine.RankingS), utils.RankingS},
		{"trends.go", "MetaTrends", new(engine.TrendS), utils.TrendS},
		{"stats.go", "MetaStats", new(engine.StatS), utils.EmptyString},
		{"thresholds.go", "MetaThresholds", new(engine.ThresholdS), utils.EmptyString},
		{"loaders.go", "MetaLoaders", new(loaders.LoaderS), utils.EmptyString},
		{"ees.go", "MetaEEs", new(ees.EeS), utils.EmptyString},
		{"analyzers.go", "MetaAnalyzer", new(analyzers.AnalyzerS), utils.EmptyString},
		{"admins.go", "MetaAdminS", new(apis.AdminSv1), utils.EmptyString},
		{"cores.go", "MetaCore", new(cores.CoreS), utils.EmptyString},
		{"guardian.go", "MetaGuardian", guardian.Guardian, utils.GuardianS},
		{"efs.go", "MetaEFs", new(efs.EfS), utils.EmptyString},
		{"ers.go", "MetaERs", new(ers.ERService), utils.ErS},
		{"tpes.go", "MetaTpes", new(tpes.TPeS), utils.EmptyString},
		// {"servicemanager.go", "MetaServiceManager", new(servmanager.ServiceManager), utils.EmptyString},
	} {
		if err := createFile(file.path, file.subsystem, file.customName, file.obj); err != nil {
			log.Fatal(err)
		}
	}

}

func createFile(filePath, subsystem, customName string, obj any) (err error) {
	var f io.WriteCloser
	if f, err = os.Create(filePath); err != nil {
		return
	}
	defer f.Close()
	return writeFile(f, subsystem, customName, obj)
}

func writeFile(w io.Writer, subsystem, customName string, obj any) (err error) {
	if _, err = w.Write([]byte(`/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

// do not modify this code because it's generated
`)); err != nil {
		return
	}
	srv, _ := engine.NewServiceWithName(obj, customName, len(customName) != 0)
	fs := token.NewFileSet()
	f := generateService(subsystem, srv)
	ast.SortImports(fs, f)
	return printer.Fprint(w, fs, f)
}

func generateService(subsystem string, srvs engine.IntService) *ast.File {
	imports := utils.NewStringSet([]string{
		"github.com/cgrates/cgrates/utils",
		"github.com/cgrates/birpc/context",
	})
	decl := make([]ast.Decl, 0)
	for k, srv := range srvs {
		if unicode.IsLetter(rune(k[len(k)-1])) {
			continue
		}
		methods := make([]string, 0, len(srv.Methods))
		for n := range srv.Methods {
			methods = append(methods, n)
		}
		sort.Strings(methods)
		for _, n := range methods {
			m := srv.Methods[n]
			decl = append(decl, generateFunc(srv.Name+n, subsystem, m.ArgType, m.ReplyType))
			imports.AddSlice(getImports(m.ArgType))
			imports.AddSlice(getImports(m.ReplyType))
		}
	}
	imports.Remove("")
	imps := make([]ast.Spec, imports.Size())
	for i, k := range imports.AsOrderedSlice() {
		imps[i] = &ast.ImportSpec{Path: &ast.BasicLit{
			// Kind:  token.STRING,
			Value: strconv.Quote(k),
		}}
	}
	decl = append([]ast.Decl{&ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: imps,
	}}, decl...)
	return &ast.File{
		Name:  ast.NewIdent("dispatchers"),
		Decls: decl,
	}
}

func generateFunc(service, subsystem string, arg, reply reflect.Type) *ast.FuncDecl {
	defer func() {
		val := recover()
		if val != nil {
			log.Println(service)
			panic(val)
		}
	}()
	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("dS")},
				Type:  &ast.StarExpr{X: ast.NewIdent("DispatcherService")},
			}},
		},
		Name: ast.NewIdent(service),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("ctx")},
					Type: &ast.StarExpr{X: &ast.SelectorExpr{
						X:   ast.NewIdent("context"),
						Sel: ast.NewIdent("Context"),
					}},
				},
				{
					Names: []*ast.Ident{ast.NewIdent("args")},
					Type:  getArgType(arg),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("reply")},
					Type:  getArgType(reply),
				},
			}},
			Results: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("err")},
				Type:  ast.NewIdent("error"),
			}}},
		},
		Body: &ast.BlockStmt{List: generateFuncBody(arg, service, subsystem)},
	}
}

type fldPath struct {
	Name      string
	IsPointer bool
}

func generatePath(arg reflect.Type, field string, kind reflect.Kind) (p []fldPath) {
	if arg.Kind() == reflect.Ptr {
		arg = arg.Elem()
	}
	if arg.Kind() != reflect.Struct {
		return
	}
	fld, has := arg.FieldByName(field)
	if !has {
		nf := arg.NumField()
		for i := 0; i < nf; i++ {
			fld := arg.Field(i)
			if fld.Type.Kind() != reflect.Struct ||
				fld.Type.Kind() != reflect.Ptr {
				continue
			}
			if p = generatePath(fld.Type, field, kind); p != nil {
				return append([]fldPath{{fld.Name, fld.Type.Kind() == reflect.Ptr}}, p...)
			}
		}
		return
	}
	p = make([]fldPath, len(fld.Index))
	cur := arg
	for i, idx := range fld.Index {
		f := cur.Field(idx)
		p[i] = fldPath{f.Name, f.Type.Kind() == reflect.Ptr}
		cur = f.Type
		if cur.Kind() == reflect.Ptr {
			cur = cur.Elem()
		}
	}
	if cur.Kind() != kind {
		return nil
	}
	return
}

func newCond(conds []*ast.BinaryExpr) *ast.BinaryExpr {
	if len(conds) == 1 {
		return conds[0]
	}
	return &ast.BinaryExpr{
		X:  conds[0],
		Op: token.LAND,
		Y:  newCond(conds[1:]),
	}
}
func generateCond(arg reflect.Type, obj, dftVal ast.Expr, field string, kind reflect.Kind) (p []ast.Stmt) {
	p = make([]ast.Stmt, 0, 2)
	p = append(p, &ast.AssignStmt{
		Lhs: []ast.Expr{obj},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{dftVal},
	})
	paths := generatePath(arg, field, kind)
	if len(paths) == 0 {
		return
	}
	paths = append([]fldPath{{"args", arg.Kind() == reflect.Ptr}}, paths...)
	conds := make([]*ast.BinaryExpr, 0, len(paths)+1)
	curPath := ""
	nilI := ast.NewIdent("nil")
	for i, p := range paths {
		if i != 0 {
			curPath += "."
		}
		curPath += p.Name
		if !p.IsPointer {
			continue
		}
		conds = append(conds, &ast.BinaryExpr{
			X:  ast.NewIdent(curPath),
			Op: token.NEQ,
			Y:  nilI,
		})
	}
	if kind == reflect.String {
		conds = append(conds, &ast.BinaryExpr{
			X:  ast.NewIdent("len(" + curPath + ")"),
			Op: token.NEQ,
			Y:  ast.NewIdent("0"),
		})
	}
	if len(conds) == 0 {
		return []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{obj},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{ast.NewIdent(curPath)},
			},
		}
	}
	p = append(p, &ast.IfStmt{
		Cond: newCond(conds),
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{obj},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{ast.NewIdent(curPath)},
			},
		}},
	})
	return
}

func generateFuncBody(arg reflect.Type, funcName, subsystem string) (p []ast.Stmt) {
	tnt := ast.NewIdent("tnt")
	p = append(p, generateCond(arg, tnt, ast.NewIdent("dS.cfg.GeneralCfg().DefaultTenant"), utils.Tenant, reflect.String)...)
	ev := ast.NewIdent("ev")
	p = append(p, generateCond(arg, ev, ast.NewIdent("make(map[string]any)"), utils.Event, reflect.Map)...)
	opts := ast.NewIdent("opts")
	p = append(p, generateCond(arg, opts, ast.NewIdent("make(map[string]any)"), "APIOpts", reflect.Map)...)

	p = append(p, &ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent("dS"),
			Sel: ast.NewIdent("Dispatch"),
		},
		Args: []ast.Expr{
			ast.NewIdent("ctx"),
			&ast.UnaryExpr{
				Op: token.AND,
				X: &ast.CompositeLit{
					Type: ast.NewIdent("utils.CGREvent"),
					Elts: []ast.Expr{
						&ast.KeyValueExpr{
							Key:   ast.NewIdent(utils.Tenant),
							Value: tnt,
						},
						&ast.KeyValueExpr{
							Key:   ast.NewIdent(utils.Event),
							Value: ev,
						},
						&ast.KeyValueExpr{
							Key:   ast.NewIdent("APIOpts"),
							Value: opts,
						},
					},
				},
			},
			&ast.SelectorExpr{
				X:   ast.NewIdent("utils"),
				Sel: ast.NewIdent(subsystem),
			},
			&ast.SelectorExpr{
				X:   ast.NewIdent("utils"),
				Sel: ast.NewIdent(funcName),
			},
			ast.NewIdent("args"),
			ast.NewIdent("reply"),
		},
	}},
	})
	return

}

func getArgType(args reflect.Type) ast.Expr {
	if name := args.Name(); len(name) != 0 {
		pkgpath := args.PkgPath()
		if len(pkgpath) == 0 {
			return ast.NewIdent(name)
		}
		return &ast.SelectorExpr{
			X:   ast.NewIdent(path.Base(pkgpath)),
			Sel: ast.NewIdent(name),
		}
	}
	switch args.Kind() {
	default:
		panic("unsuported argument")
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64,
		reflect.Complex128, reflect.String:
		return ast.NewIdent(args.Name())
	case reflect.Interface:
		name := args.Name()
		if len(name) == 0 {
			name = "any"
		}
		return ast.NewIdent(name)
	case reflect.Ptr:
		return &ast.StarExpr{
			X: getArgType(args.Elem()),
		}
	case reflect.Struct:
		pkgpath := args.PkgPath()
		if len(pkgpath) == 0 {
			return ast.NewIdent(args.Name())
		}
		return &ast.SelectorExpr{
			X:   ast.NewIdent(path.Base(pkgpath)),
			Sel: ast.NewIdent(args.Name()),
		}
	case reflect.Array, reflect.Slice:
		return &ast.ArrayType{
			Elt: getArgType(args.Elem()),
		}
	case reflect.Map:
		return &ast.MapType{
			Key:   getArgType(args.Key()),
			Value: getArgType(args.Elem()),
		}
	}
}

func getImports(args reflect.Type) []string {
	switch args.Kind() {
	default:
		panic("unsuported argument")
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64,
		reflect.Complex128, reflect.String, reflect.Interface:
		return nil
	case reflect.Ptr, reflect.Array, reflect.Slice:
		return append(getImports(args.Elem()), args.PkgPath())
	case reflect.Struct:
		return []string{args.PkgPath()}
	case reflect.Map:
		args.PkgPath()
		key := append(getImports(args.Key()), args.PkgPath())
		return append(key, getImports(args.Elem())...)
	}
}
