package main

import (
	"encoding/json"
	"html/template"
	"lamprey/core"
	"lamprey/core/db"
	"lamprey/core/deployer"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
)

type App struct {
	PageManager db.PageManager
	Config      core.Config
	Deployer    deployer.Deployer
}

func main() {
	a := core.ReadConfig("config.toml")
	pm, err := db.NewPageManager("db.sqlite")
	if err != nil {
		panic(err)
	}
	app := App{
		PageManager: *pm,
		Config:      a,
		Deployer:    deployer.DummyDeployer{},
	}

	if a.DeployToFolder != nil {
		app.Deployer = deployer.InitFsDeployer(*a.DeployToFolder)
	}

	// start server
	if a.Http != nil {
		httpServer(*a.Http, app)
	}

	if a.FastCgi != nil {
		fastCgiServer(*a.FastCgi, app)
	}

	if _, ok := app.Deployer.(deployer.DummyDeployer); ok {
		panic("No deployer configured!")
	}
}

func httpServer(config core.HttpConfig, app App) {
	println("Starting http server on " + config.Address)

	if app.Config.DeployToFolder != nil {
		http.ListenAndServe(config.Address, getHttpHandler("/lamprey", app))
	} else {
		http.ListenAndServe(config.Address, getHttpHandler("", app))
	}
}

func fastCgiServer(config core.FastCgiConfig, app App) {
	println("Starting fastcgi server on " + config.Address)

	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		println("Unable to bind to address:" + config.Address)
		panic(err)
	}
	err = fcgi.Serve(listener, getHttpHandler("", app))
	if err != nil {
		println("Unable to start fastcgi server")
		panic(err)
	}

}

func getHttpHandler(prefix string, app App) http.Handler {
	a := http.ServeMux{}
	a.HandleFunc(prefix+"/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Only admins should be able to access this page"))
		w.Write([]byte(r.RequestURI))
	})
	a.HandleFunc("GET "+prefix+"/edit/article/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		page, err := app.PageManager.GetPage(r.PathValue("objectid"))
		if err != nil {
			w.Write([]byte("Unable to get page"))
			panic(err)
		}

		templates, err := template.ParseFiles("views/layout.html.gotmpl", "views/edit.html.gotmpl")
		if err != nil {
			http.Error(w, "Error loading templates", http.StatusInternalServerError)
			log.Println("Error parsing templates:", err)
			return
		}

		// Render the template with the provided Page data
		err = templates.ExecuteTemplate(w, "layout.html.gotmpl", page)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Println("Error executing template:", err)
		}
	})
	a.HandleFunc("POST "+prefix+"/edit/article/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		objectid := r.PathValue("objectid")
		page, err := app.PageManager.GetPage(objectid)
		if err != nil {
			w.Write([]byte("Unable to get page"))
			panic(err)
		}
		page.Content = r.FormValue("content")
		err = app.PageManager.UpdatePage(page.ID, page.Title, page.Content, page.Data)
		if err != nil {
			w.Write([]byte("Unable to update page"))
			panic(err)
		}
		err = app.Deployer.DeployArticle(*page)
		if err != nil {
			w.Write([]byte("Unable to deploy article"))
			panic(err)
		}

		http.Redirect(w, r, prefix+"/edit/article/"+objectid, http.StatusSeeOther)
	})
	a.HandleFunc("GET "+prefix+"/data/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Editing data for " + r.PathValue("objectid")))
		w.Write([]byte(r.RequestURI))
	})
	a.HandleFunc("POST "+prefix+"/data/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Saving data for " + r.PathValue("objectid")))
		w.Write([]byte(r.RequestURI))

		page, err := app.PageManager.GetPage(r.PathValue("objectid"))
		if err != nil {
			w.Write([]byte("Unable to get page"))
			panic(err)
		}
		json.Unmarshal([]byte(r.FormValue("data")), &page.Data)
		err = app.PageManager.UpdatePage(page.ID, page.Title, page.Content, page.Data)
		if err != nil {
			w.Write([]byte("Unable to update page"))
			panic(err)
		}
		err = app.Deployer.DeployData(*page)
		if err != nil {
			w.Write([]byte("Unable to deploy article"))
			panic(err)
		}
	})
	a.HandleFunc("GET "+prefix+"/new", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Creating new entry"))
		w.Write([]byte(r.RequestURI))
	})
	a.HandleFunc("PUT "+prefix+"/new", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Creating new entry"))
		w.Write([]byte(r.RequestURI))
	})
	if app.Config.DeployToFolder != nil {
		println("Staticly serving from folder: " + app.Config.DeployToFolder.Path)
		a.Handle("/", http.FileServer(http.Dir(app.Config.DeployToFolder.Path)))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.ServeHTTP(w, r)
	})
}
