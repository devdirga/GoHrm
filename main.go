package main

import (
	_ "creativelab/ecleave-dev/webext"
	"log"
	"net/http"

	"github.com/creativelab/knot/knot.v1"
)

func main() {

	app := knot.GetApp("ecleave")
	if app == nil {
		log.Println("App not found....")
		return
	}
	// routes := make(map[string]knot.FnContent, 1)
	// routes["/"] = func(k *knot.WebContext) interface{} {
	// 	http.Redirect(k.Writer, k.Request, "/login/default", http.StatusTemporaryRedirect)
	// 	return true
	// },
	// "prerequest":func(k *knot.WebContect) interface{}{
	// 	if k.Request.URL.String() == "/mail/linked" && k.Session("username") == nil{

	// 	}
	// }

	routes := map[string]knot.FnContent{
		"/": func(k *knot.WebContext) interface{} {
			http.Redirect(k.Writer, k.Request, "/login/default", http.StatusTemporaryRedirect)
			return true
		},
		"prerequest": func(k *knot.WebContext) interface{} {
			if k.Request.URL.String() == "/mail/ResponseLeader" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/mail/ResponseLeader", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/mail/ResponseBA" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/mail/ResponseBA", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/mail/ResponseLeaderDecline" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/mail/ResponseLeaderDecline", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/mail/Responsebadecline" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/mail/Responsebadecline", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/mail/ResponseManager" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/mail/ResponseManager", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/mail/ResponseManagerDecline" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/mail/ResponseManagerDecline", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/mail/ResetPassword" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/mail/resetpassword", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/overtime/UserOvertime" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/overtime/UserOvertime", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/overtime/ApprovedOvertime" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/overtime/ApprovedOvertime", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/overtime/DeclinedOvertime" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/overtime/DeclinedOvertime", http.StatusTemporaryRedirect)
				return true
			} else if k.Request.URL.String() == "/overtime/ManagerOvertime" && k.Session("username") == nil {
				http.Redirect(k.Writer, k.Request, "/overtime/ManagerOvertime", http.StatusTemporaryRedirect)
				return true
			} //else if k.Request.URL.String() == "/mail/admapprovecancelleave" && k.Session("username") == nil {
			// 	http.Redirect(k.Writer, k.Request, "/mail/admapprovecancelleave", http.StatusTemporaryRedirect)
			// 	return true
			// } else if k.Request.URL.String() == "/mail/admdeclinecancelleave" && k.Session("username") == nil {
			// 	http.Redirect(k.Writer, k.Request, "/mail/admdeclinecancelleave", http.StatusTemporaryRedirect)
			// 	return true
			// }
			return nil
		},
	}
	knot.StartAppWithFn(app, "localhost:8078", routes)
}
