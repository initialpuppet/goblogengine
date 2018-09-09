package blog

import (
	"goblogengine/middleware/auth"
	"goblogengine/middleware/basehandler"
	"goblogengine/middleware/flashes"

	"goblogengine/external/github.com/gorilla/mux"
)

// Init accepts an HTTP router and sets the routes for the public facing blog pages
func Init(r *mux.Router) {
	r.HandleFunc("/", basehandler.MakeHandler(auth.AddInfo(HomeGET)))

	r.HandleFunc("/page/{pagenumber}", basehandler.MakeHandler(auth.AddInfo(HomeGET)))
	r.HandleFunc("/post/{postslug}", basehandler.MakeHandler(auth.AddInfo(PostGET)))
	r.HandleFunc("/image/{imageid}", basehandler.MakeHandler(auth.AddInfo(ServeImageGET)))
	r.HandleFunc("/atom", AtomGET)

	r.HandleFunc("/admin", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminHomeGET))))).Methods("GET")

	r.HandleFunc("/admin/post/list", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostListGET))))).Methods("GET")
	r.HandleFunc("/admin/post/add", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostEditGET))))).Methods("GET")
	r.HandleFunc("/admin/post/add", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostEditPOST))))).Methods("POST")

	r.HandleFunc("/admin/post/edit/{postslug}", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostEditGET))))).Methods("GET")
	r.HandleFunc("/admin/post/edit/{postslug}", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostEditPOST))))).Methods("POST")
	r.HandleFunc("/admin/post/publish", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostPublishPOST))))).Methods("POST")
	r.HandleFunc("/admin/post/unpublish", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostUnpublishPOST))))).Methods("POST")
	r.HandleFunc("/admin/post/delete", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPostDeletePOST))))).Methods("POST")
	r.HandleFunc("/admin/post/preview/{postslug}/{version}", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminPreviewPostVersionGET))))).Methods("GET")

	r.HandleFunc("/admin/author/list", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminAuthorListGET))))).Methods("GET")
	r.HandleFunc("/admin/author/add", basehandler.MakeHandler(auth.AddInfo(flashes.Add(AdminAuthorInsertGET)))).Methods("GET")
	r.HandleFunc("/admin/author/add", basehandler.MakeHandler(auth.AddInfo(flashes.Add(AdminAuthorInsertPOST)))).Methods("POST")

	r.HandleFunc("/admin/image/list", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminImageListGET))))).Methods("GET")
	r.HandleFunc("/admin/image/list.json", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminImageListJSGET))))).Methods("GET")
	r.HandleFunc("/admin/image/list", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminImageListPOST))))).Methods("POST")
	r.HandleFunc("/admin/image/update", basehandler.MakeHandler(auth.AddInfo(auth.Require(AdminImageUpdateJSPOST)))).Methods("POST")
	r.HandleFunc("/admin/image/upload", basehandler.MakeHandler(auth.AddInfo(auth.Require(AdminImageUploadJSPOST)))).Methods("POST")
	r.HandleFunc("/admin/image/delete", basehandler.MakeHandler(auth.AddInfo(auth.Require(AdminImageDeletePOST)))).Methods("POST")
	r.HandleFunc("/admin/image/deleteall", basehandler.MakeHandler(auth.AddInfo(auth.Require(AdminImageDeleteAllPOST)))).Methods("POST")

	r.HandleFunc("/admin/category/list", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(CategoryListGET))))).Methods("GET")
	r.HandleFunc("/admin/category/list", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(CategoryListPOST))))).Methods("POST")
	r.HandleFunc("/admin/category/delete", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(CategoryDeletePOST))))).Methods("POST")

	r.HandleFunc("/admin/data", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminDataGET))))).Methods("GET")
	r.HandleFunc("/admin/data", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminImportPostsPOST))))).Methods("POST")
	r.HandleFunc("/admin/data/export", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminExportPostsPOST))))).Methods("POST")

	r.HandleFunc("/admin/reset", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminResetGET))))).Methods("GET")
	r.HandleFunc("/admin/reset", basehandler.MakeHandler(auth.AddInfo(auth.Require(flashes.Add(AdminResetPOST))))).Methods("POST")

	r.NotFoundHandler = basehandler.MakeHandler(NotFound)
}
