package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

//

var db_mutex = &sync.Mutex{}

//

func createUrlStatusTable(db *sql.DB) {
	createURLTableSQL := `CREATE TABLE url_status (
		"id"              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"digest_num"      TEXT,
		"line_in_digest"  TEXT,
        "valid"           BOOL,
        "timestamp"       INT,
		"orig_url"        TEXT,
		"derived_url"     TEXT
	  );` // SQL Statement for Create Table

	fmt.Println("Create url_status table...")
	statement, err := db.Prepare(createURLTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal("Prepare crateURLTableSQL failed: " + err.Error())
	}
	defer statement.Close()
	statement.Exec() // Execute SQL Statements
	fmt.Println("url_status table created")
}

func insertUrlStatus(db *sql.DB, digest_num string, line_in_digest int, valid bool, orig_url string, derived_url string) {
	//fmt.Println("Inserting URL Status ...")
	insertUrlStatusSQL := `INSERT INTO url_status(digest_num, line_in_digest, valid, timestamp, orig_url, derived_url) VALUES (?, ?, ?, ?, ?, ?)`
	db_mutex.Lock()
	statement, err := db.Prepare(insertUrlStatusSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln("Prepare INSERT INTO failed: " + err.Error())
	}
	timestamp := time.Now()
	fmt.Printf("%s INSERT %s  %3d  %t   %s   %s   %s\n", digest_num, digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	_, err = statement.Exec(digest_num, line_in_digest, valid, timestamp, orig_url, derived_url)
	db_mutex.Unlock()
	if err != nil {
		log.Printf("Execute INSERT INTO failed: " + err.Error())
		time.Sleep(500 * time.Second)
	}
	statement.Close()
}

// CheckRedirect specifies the policy for handling redirects.
// If CheckRedirect is not nil, the client calls it before
// following an HTTP redirect. The arguments req and via are
// the upcoming request and the requests made already, oldest
// first. If CheckRedirect returns an error, the Client's Get
// method returns both the previous Response (with its Body
// closed) and CheckRedirect's error (wrapped in a url.Error)
// instead of issuing the Request req.
// As a special case, if CheckRedirect returns ErrUseLastResponse,
// then the most recent response is returned with its body
// unclosed, along with a nil error.
//
// If CheckRedirect is nil, the Client uses its default policy,
// which is to stop after 10 consecutive requests.

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	//fmt.Printf("redirectPolicyFunc: Request Method = %s\n", req.Method)
	n_redirects := len(via)
	fmt.Printf("redirectPolicyFunc: %s  -->  %s\n", via[n_redirects-1].URL, req.URL)
	//for i, prev_req := range via {
	//    fmt.Printf("redirectPolicyFunc: prev URL[%d]            = %s\n", i, prev_req.URL)
	//}

	if n_redirects > 10 {
		reason := fmt.Sprintf("Too Many Redirects for %s", via[0].URL)
		return errors.New(reason)
	} else {
		return nil
	}
}

func display_num_open_files(thread_num int, digest_num string, reason string) {

	fid, err := syscall.Dup(1)
	if err != nil {
		fmt.Printf("%s Thread %3d %s has TOO MANYd files open\n", digest_num, thread_num, reason)
		return
	}

	fmt.Printf("%s Thread %3d %s has %d files open\n", digest_num, thread_num, reason, fid)

	err = syscall.Close(fid)
	if err != nil {
		fmt.Printf("%s Thread %3d %s could not close %d\n", digest_num, thread_num, reason, fid)
		return
	}
}

func check_url(client *http.Client, thread_num int, db *sql.DB, depth int, digest_num string, url string) {

	display_num_open_files(thread_num, digest_num, "entering check_url")
	if false {
		log.Printf("%s Thread %3d Testing Insert without network activity", digest_num, thread_num)
		insertUrlStatus(db, digest_num, 1, false, url, "Testing Insert without network activity")
		return
	}
	// we really should masquerade as: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:84.0) Gecko/20100101 Firefox/84.0"
	// currently we arrive as:  "Go-http-client/1.1"

	//resp, err := client.Head(url)

	req, _ := http.NewRequest("GET", url, nil)
	display_num_open_files(thread_num, digest_num, "after NewRequest")

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:84.0) Gecko/20100101 Firefox/84.0")

	resp, err := client.Do(req) // WARNING -- this consumes a file descriptor that must be closed!
	//    defer resp.Close() Close doesn't exist.  Grrrr
	display_num_open_files(thread_num, digest_num, "after client.Do")

	//fmt.Printf("check_url(\"%s\",  \"%s\")\n", digest_num, url)
	if err != nil {
		//fmt.Printf("Error fetching \"%s\": %s\n", url, err)
		err_string := fmt.Sprintf("%s", err)
		if strings.Contains(err_string, "unsupported protocol scheme") {
			//fmt.Printf ("Trying adding http://\n");
			if depth < 1 {
				check_url(client, thread_num, db, depth+1, digest_num, "http://"+url)
			} else {
				fmt.Printf("%s check_url() depth to deep for %s\n", digest_num, url)
				line_in_digest := 1 // kludge
				orig_url := url
				insertUrlStatus(db, digest_num, line_in_digest, false, orig_url, "excess recursion")
			}
		} else { // other than unsupported protocol.  No retry
			fmt.Printf("%s Failed for %s : %s\n", digest_num, url, err_string)
			line_in_digest := 1 // kludge
			orig_url := url
			derived_url := url
			insertUrlStatus(db, digest_num, line_in_digest, false, orig_url, derived_url)
		}
	} else {
		var derived_url string
		// got a non-fatal response
		resp_url := resp.Request.URL
		fmt.Printf("%s req url              = %s\n", digest_num, url)
		//fmt.Printf("resp.URL.Scheme      = %s\n", resp_url.Scheme)
		//fmt.Printf("resp.URL.Opaque      = %s\n", resp_url.Opaque)
		//fmt.Printf("resp.URL.Host        = %s\n", resp_url.Host)
		//fmt.Printf("resp.URL.Path        = %s\n", resp_url.Path)
		//fmt.Printf("resp.URL.RawPath     = %s\n", resp_url.RawPath)
		//fmt.Printf("resp.URL.RawQuery    = %s\n", resp_url.RawQuery)
		//fmt.Printf("resp.URL.Fragment    = %s\n", resp_url.Fragment)
		//fmt.Printf("resp.URL.RawFragment = %s\n", resp_url.RawFragment)

		final_url := fmt.Sprintf("%s://%s%s", resp_url.Scheme, resp_url.Host, resp_url.Path)
		if len(resp_url.RawQuery) > 0 {
			final_url += "?" + resp_url.RawQuery
		}
		if final_url != url {
			fmt.Printf("%s resp derived_url     = %s\n", digest_num, final_url)
		}

		//fmt.Printf ("Result from \"%s\" is:\n\"%s\"\n", url, resp)
		fmt.Printf("%s Status %d for %s\n", digest_num, resp.StatusCode, url)
		line_in_digest := 1 // kludge
		orig_url := url
		if final_url == orig_url {
			derived_url = ""
		} else {
			derived_url = final_url
		}
		insertUrlStatus(db, digest_num, line_in_digest, true, orig_url, derived_url)
	}
}

func url_thread(thread_num int, db *sql.DB, work, done chan string, httpClient *http.Client) {
	var each_ln string
	//
	// we are going to rate-limit each thread to 1 per second
	//
	// implement at first with time.Sleep of 1 second, but
	// eventually get current time and check time of completion
	// to decide how long to sleep

	for true {
		each_ln = <-work
		//fmt.Printf("url_thread_%02d got %s\n", thread_num, each_ln)
		// time.Sleep(100 * time.Millisecond)
		fields := strings.Split(each_ln, "\t")
		check_url(httpClient, thread_num, db, 0, fields[2], fields[3])
		done <- fmt.Sprintf("url_thread_%02d done with %s", thread_num, each_ln)
		// time.Sleep(100*time.Millisecond)
		fmt.Printf("thread %3d done sleeping\n", thread_num)
	}
}

func main() {
	err := os.Remove("sqlite-database.db") // I delete the file to avoid duplicated records.  // SQLite is a file based database.
	if err != nil {
		log.Printf("Failed delete " + err.Error())
	}

	fmt.Println("Creating sqlite-database.db...")
	dbfile, err := os.Create("sqlite-database.db") // Create SQLite file
	if err != nil {
		log.Fatal("Create database file failed: " + err.Error())
	}
	dbfile.Close()
	fmt.Println("sqlite-database.db created")

	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
	defer sqliteDatabase.Close()                                     // Defer Closing the database
	createUrlStatusTable(sqliteDatabase)                             // Create Database Tables

	// os.Open() opens specific file in
	// read-only mode and this return
	// a pointer of type os.
	file, err := os.Open("sample_urls.txt")

	if err != nil {
		log.Fatalf("Open saved_urls.txt: failed to open: " + err.Error())

	}

	// The bufio.NewScanner() function is called in which the
	// object os.File passed as its parameter and this returns a
	// object bufio.Scanner which is further used on the
	// bufio.Scanner.Split() method.
	scanner := bufio.NewScanner(file)

	// The bufio.ScanLines is used as an
	// input to the method bufio.Scanner.Split()
	// and then the scanning forwards to each
	// new line using the bufio.Scanner.Scan()
	// method.
	scanner.Split(bufio.ScanLines)
	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	// The method os.File.Close() is called
	// on the os.File object to close the file
	file.Close()

	// and then a loop iterates through
	// and prints each of the slice values.

	work := make(chan string, 1000)
	done := make(chan string, 100)

	httpClient := &http.Client{
		CheckRedirect: redirectPolicyFunc,
	}

	for i := 0; i < 1; i++ {
		go url_thread(i+1, sqliteDatabase, work, done, httpClient)
	}

	work_queue_len := 0
	num_sent := 0
	num_rcvd := 0

	for _, each_ln := range text {
		//fields := strings.Split(each_ln, "\t")
		//fmt.Printf ("Sending to chan: %s\n", each_ln)
		work <- each_ln
		work_queue_len += 1
		num_sent += 1
		if work_queue_len == 100 {
			result := <-done
			work_queue_len -= 1
			if false {
				fmt.Printf("Result rcvd A:  %s\n", result)
			}
			num_rcvd++
		}
	}

	for work_queue_len > 0 {
		result := <-done
		work_queue_len -= 1
		if false {
			fmt.Printf("Result rcvd B:  %s\n", result)
		}
		num_rcvd++
	}
	// really need to wait until all are done...
	fmt.Printf("All work completed, sent %d, rcvd %d\n", num_sent, num_rcvd)
}
