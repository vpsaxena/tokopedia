package main

import "net/http"
import "database/sql"						//package for sql manupulation
import	_ "database/sql/driver/mysql"       //driver
import "log"
import "fmt"
import "strings"
import "net/smtp"	
import "time"
import b64 "encoding/base64"	

var db *sql.DB
var row *sql.Rows

func main() {
  fmt.Println("application started......")
  http.HandleFunc("/signup",signupHandler)
  http.HandleFunc("/login",loginHandler)
  http.HandleFunc("/profile",profileHandler)
  http.HandleFunc("/updating",updateHandler)
  http.HandleFunc("/active",active)
  http.HandleFunc("/",welcome)
  http.ListenAndServe(":8080", nil)
}

func active(res http.ResponseWriter, req *http.Request){
	
	fmt.Println("loading activation page.....")
	
	email:=req.FormValue("id")
	
	//get database connection
	fmt.Println("Opening connection......")
    db,err_open:=sql.Open("mysql","root:sqla@tcp(127.0.0.1:3306)/tokopedia")
    fmt.Println("Connection opened......")
    if err_open!=nil {
      log.Fatal(err_open)
    }
	
	
	//fire update query
	fmt.Println("Inserting in db .....")
    _, err_exec := db.Exec("UPDATE users set active=? where mailid=?","true",email)
    fmt.Println("Inserted in db......")
    if err_exec!=nil {
      log.Fatal(err_exec)
      fmt.Println(err_exec)
	  fmt.Fprintf(res,"<h1>Some error occurred, please try again later.</h1>")
    } else {
	  fmt.Fprintf(res,"<h1>Account has been successfully activated.</h1>")
	}
}

func welcome(res http.ResponseWriter, req *http.Request) {
  if req.Method!="POST" {
    http.ServeFile(res,req,"login.html")
	fmt.Println("Calling login.html.....")
	return
  }
}

func updateHandler(res http.ResponseWriter, req *http.Request) {
  
  fmt.Println("Updating......")
  
  fname := req.FormValue("fname")
  lname := req.FormValue("lname")
  dob := req.FormValue("dob")
  fmt.Println("Value retrieved from form........")
  
  ck_mail,_:=req.Cookie("mailid")
  
  //open connection of mysql
  fmt.Println("Opening connection......")
  db,err_open:=sql.Open("mysql","root:sqla@tcp(127.0.0.1:3306)/tokopedia")
  fmt.Println("Connection opened......")
  if err_open!=nil {
      log.Fatal(err_open)
  }
  
  //updating in mysql
  fmt.Println("Inserting in db .....")
  _, err_exec := db.Exec("UPDATE users set fname=?,lname=?,dob=? where mailid=?",fname,lname,dob,ck_mail.Value)
  fmt.Println("Inserted in db......")
  if err_exec!=nil {
    log.Fatal(err_exec)
    fmt.Println(err_exec)
	fmt.Fprintf(res,"<h1>Some error occurred, please try again later.</h1>")
  } else {
    fmt.Println("setting data done.")
	http.Redirect(res,req,"profile",303)
  }
}

func profileHandler(res http.ResponseWriter, req *http.Request) {
  if req.Method!="POST" {
    http.ServeFile(res,req,"profile.html")
	fmt.Println("caling profile.html ......",req.Method)
	return
  }
 
  http.Redirect(res,req,"updating",303)
  
  del_cookie1 := &http.Cookie(Name:"fname",Value="")
  del_cookie2 := &http.Cookie(Name:"lname",Value="")
  del_cookie3 := &http.Cookie(Name:"dob",Value="")
  del_cookie4 := &http.Cookie(Name:"mail",Value="")
  del_cookie5 := &http.Cookie(Name:"pwd",Value="")
  
  http.SetCookie(del_cookie1)
  http.SetCookie(del_cookie2)
  http.SetCookie(del_cookie3)
  http.SetCookie(del_cookie4)
  http.SetCookie(del_cookie5)
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
  uid := req.FormValue("uid")
  upass := req.FormValue("upass")
  
  var fname,lname,dob,mailid,pwd,active string
  var count int
  count=0
  
  //open connection of mysql
  fmt.Println("Opening connection......")
  db,err_open:=sql.Open("mysql","root:sqla@tcp(127.0.0.1:3306)/tokopedia")
  fmt.Println("Connection opened......")
  if err_open!=nil {
      log.Fatal(err_open)
  }
  
  //get data from db
  rows, err_get:=db.Query("select fname,lname,dob,mailid,pwd,active from users where mailid='"+uid+"'")
  if err_get!=nil {
      log.Fatal(err_get)
  }
  
  //checking for user
  if rows.Next() {
	rows.Scan(&fname,&lname,&dob,&mailid,&pwd,&active)
	fmt.Println("Uid: "+uid+"\nDBUid: "+mailid+"")
    if (strings.Compare(mailid, uid)==0) {
	  fmt.Println("Found user.")
	  fmt.Println("pwd="+pwd)
	  pd,_:=b64.URLEncoding.DecodeString(pwd)
	  //fmt.Println("decoded pwd="+pd)
	  pwd_decoded:=string(pd)
	  fmt.Println("decoded password="+pwd_decoded)
	  if (strings.Compare(upass, pwd_decoded)==0) {
	    fmt.Println("Password Match.")
	    if (active=="false") {
		  fmt.Println("Not active.")
		  http.ServeFile(res,req,"login.html")
		  fmt.Println("Kindly click the link that has been mailed to your ID to activate your account.")
	    } else {
		    fmt.Println("Active.")
	        //http.ServeFile(res,req,"profile.html")
			exp:=time.Now().Add(365 * 24 * time.Hour)
			ck_fname:=http.Cookie{Name:"fname",Value:fname,Expires:exp}
			http.SetCookie(res,&ck_fname)
			ck_lname:=http.Cookie{Name:"lname",Value:lname,Expires:exp}
			http.SetCookie(res,&ck_lname)
			ck_dob:=http.Cookie{Name:"dob",Value:dob,Expires:exp}
			http.SetCookie(res,&ck_dob)
			ck_mail:=http.Cookie{Name:"mail",Value:mailid,Expires:exp}
			http.SetCookie(res,&ck_mail)
			ck_pwd:=http.Cookie{Name:"pwd",Value:pwd,Expires:exp}
			http.SetCookie(res,&ck_pwd)
			http.Redirect(res,req,"profile",303)
	    }
	  } else {
	      fmt.Println("Password is wrong.")
		  http.ServeFile(res,req,"login.html")
	  }
	  count++
	}
  }
  
  //when no user found
  if count==0 {
    fmt.Println("No user found. Please sign up.")
	http.ServeFile(res,req,"login.html")
  }
}


func signupHandler(res http.ResponseWriter, req *http.Request) {
  //connect with html page
  if req.Method!="POST" {
    http.ServeFile(res,req,"signup.html")
	fmt.Println("caling signup.html ......",req.Method)
	return
  }
  
  var mailid string
 
  //taking value from form
  fname := req.FormValue("fname")
  lname:=req.FormValue("lname")
  dob:=req.FormValue("dob")
  mail:=req.FormValue("mail")
  pwd:=b64.StdEncoding.EncodeToString([]byte(req.FormValue("pwd")))
  fmt.Println("Value retrieved from form........")
  
  
  //open connection of mysql
  fmt.Println("Opening connection......")
  db,err_open:=sql.Open("mysql","root:sqla@tcp(127.0.0.1:3306)/tokopedia")
  fmt.Println("Connection opened......")
  if err_open!=nil {
      log.Fatal(err_open)
  }
  
  //get data from db
  rows, err_get:=db.Query("select mailid from users where mailid='"+mail+"'")
  if err_get!=nil {
      log.Fatal(err_get)
  }
  
  if rows.Next() {
      rows.Scan(&mailid)
	  if (strings.Compare(mailid,mail)==0) {
	      fmt.Println("User already exist.")
	    }
  } else {
      //inserting in mysql
      fmt.Println("Inserting in db .....")
      _, err_exec := db.Exec("INSERT INTO users(fname,lname,dob,mailid,pwd,active,active_date) VALUES(?,?,?,?,?,?,?)",fname,lname,dob,mail,pwd,"false",time.Now().Format("2006-2-1"))
      fmt.Println("Inserted in db......")
      if err_exec!=nil {
        log.Fatal(err_exec)
	    fmt.Println(err_exec)
	    http.ServeFile(res,req,"signup.html")
      } else {
	    //sending mail
        auth:=smtp.PlainAuth("","your.dairy.information@gmail.com","your.information","smtp.gmail.com")
        to:=[]string{mail}
        msg:=[]byte("Hi, Please click here to activate your account.\n\nhttp://localhost:8080/active?id=vrindasaxena@ymail.com")
        err_mail:=smtp.SendMail("smtp.gmail.com:587",auth,"vrindasaxena3@gmail.com",to,msg)
        if err_mail!=nil {
		  log.Fatal(err_mail)
	    } else {
		  fmt.Println("Mail sent successfully")
	    }
	    fmt.Println("Registration successful.")
        http.ServeFile(res,req,"success_signup.html")
	  }
  }
  
  //closing connections
  defer row.Close()
  defer db.Close()
}