package quest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// Subversion
func Subversion(db *gorm.DB, data []map[string]interface{}, ischeck bool) {
	for _, team := range data {
		go func(t map[string]interface{}) {
			resp, err := request(
				http.MethodGet,
				fmt.Sprintf("http://%s/listing.php", t["ip"]),
				t["hostname"].(string),
				nil, nil)
			if err != nil {
				log.Println("[-]", err) // cancel caught
				healthcheck(db, quest1, t["id"].(int), ischeck, false)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				// read body
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Println("[-]", err)
					healthcheck(db, quest1, t["id"].(int), ischeck, false)
					return
				}

				// check keywords
				healthcheck(db, quest1, t["id"].(int), ischeck, strings.Contains(string(body), "WebSVN"))
			} else {
				healthcheck(db, quest1, t["id"].(int), ischeck, false)
			}
		}(team)
	}
}
