package main

// https://www.accesstomemory.org/en/docs/2.3/user-manual/import-export/csv-import/#csv-import-repositories

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	base     = "http://daa.webfactional.com/"
	start    = "archives/22"
	headings = "identifier,uploadLimit,authorizedFormOfName,contactPerson,streetAddress,telephone,email,fax,website,geoculturalContext,internalStructures,collectingPolicies,holdings,findingAids,openingTimes,accessConditions,disabledAccess,researchServices,reproductionServices,publicFacilities,maintenanceNote,descInstitutionIdentifier,descRevisionHistory,descSources,culture"
)

type Entry struct {
	Name    string
	URL     string
	About   *Values
	Contact *Values
}

func (e *Entry) ID() string {
	return strings.Replace(e.URL, "/archives", "daa", 1)
}

func (e *Entry) UploadLimit() string {
	return "0"
}

func (e *Entry) AuthorizedFormofName() string {
	return e.Name
}

func (e *Entry) ContactPerson() string {
	if str, ok := e.Contact.get("Enquiries to"); ok {
		return str
	}
	return e.About.mustGet("Officer in charge")
}

func (e *Entry) StreetAddress() string {
	street, sok := e.Contact.get("Street address")
	post, pok := e.Contact.get("Postal address")
	if (sok && !pok) || street == post {
		return street
	} else if pok && !sok {
		return post
	}
	return street + "\n\n" + "Postal address:\n" + post
}

func (e *Entry) Telephone() string {
	return e.Contact.mustGet("Phone")
}

func (e *Entry) Email() string {
	return e.Contact.mustGet("Email")
}

func (e *Entry) Fax() string {
	return e.Contact.mustGet("Fax")
}

func (e *Entry) Website() string {
	return e.Contact.mustGet("Website")
}

func (e *Entry) InternalStructures() string {
	ret := []string{}
	if str, ok := e.About.get("Officer in charge"); ok {
		ret = append(ret, "Officer in charge: "+str)
	}
	if str, ok := e.Contact.get("Note"); ok {
		ret = append(ret, "Note: "+str)
	}
	if str, ok := e.About.get("See also"); ok {
		ret = append(ret, "See also: "+str)
	}
	return strings.Join(ret, "\n\n")
}

func (e *Entry) CollectingPolicies() string {
	return e.About.mustGet("Acquisition focus")
}

func (e *Entry) Holdings() string {
	ret := []string{}
	if str, ok := e.About.get("Major holdings"); ok {
		ret = append(ret, str)
	}
	if str, ok := e.About.get("Quantity"); ok {
		ret = append(ret, "Quantity: "+str)
	}
	if str, ok := e.About.get("References"); ok {
		ret = append(ret, "References: "+str)
	}
	return strings.Join(ret, "\n\n")
}

func (e *Entry) FindingAids() string {
	return e.About.mustGet("Guides")
}

func (e *Entry) OpeningTimes() string {
	return e.About.mustGet("Hours & facilities")
}

func (e *Entry) AccessConditions() string {
	return e.About.mustGet("Access")
}

//maintenanceNote,descInstitutionIdentifier,descRevisionHistory,descSources,culture"

func ToCSV(e []*Entry) [][]string {
	maintenanceNote := fmt.Sprintf("This entry was imported from the legacy Directory of Australian Archives on %s", time.Now().Format("Jan 2 2006"))
	firstRow := strings.Split(headings, ",")
	ret := make([][]string, 1, 1000)
	ret[0] = firstRow
	for _, v := range e {
		row := make([]string, len(firstRow))
		row[0] = v.ID()
		row[1] = v.UploadLimit()
		row[2] = v.AuthorizedFormofName()
		row[3] = v.ContactPerson()
		row[4] = v.StreetAddress()
		row[5] = v.Telephone()
		row[6] = v.Email()
		row[7] = v.Fax()
		row[8] = v.Website()
		// row[9] = geoculturalContext
		row[10] = v.InternalStructures()
		row[11] = v.CollectingPolicies()
		row[12] = v.Holdings()
		row[13] = v.FindingAids()
		row[14] = v.OpeningTimes()
		row[15] = v.AccessConditions()
		// row[16] disabledAccess
		// row[17] researchServices
		// row[18] reproductionServices
		// row[19] publicFacilities
		row[20] = maintenanceNote
		// row[21] descInstitutionIdentifier
		// row[22] descRevisionHistory
		// row[23] descSources
		// row[24] culture
		ret = append(ret, row)
	}
	return ret
}

type Values [][2]string

func (v *Values) get(k string) (string, bool) {
	for _, w := range *v {
		if w[0] == k {
			return w[1], true
		}
	}
	return "", false
}

func (v *Values) mustGet(k string) string {
	str, _ := v.get(k)
	return str
}

var repl = strings.NewReplacer("\\\"", "\\\"", "\"", "\\\"", "\\n", "\\n", "\\t", "\\t", "\n", "\\n", "\t", "\\t")

func (v *Values) String() string {
	buf := &bytes.Buffer{}
	buf.WriteByte('[')
	for i, w := range *v {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, "{\"%s\":\"%s\"}", w[0], w[1])
	}
	buf.WriteByte(']')
	return string(buf.Bytes())
}

func (v *Values) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte('[')
	for i, w := range *v {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, "{\"%s\":\"%s\"}", repl.Replace(w[0]), repl.Replace(w[1]))
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

func (v *Values) UnmarshalJSON(b []byte) error {
	buf := bytes.NewBuffer(b)
	dec := json.NewDecoder(buf)
	var key string
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		str, ok := t.(string)
		if !ok {
			continue
		}
		if key == "" {
			key = str
		} else {
			*v = append(*v, [2]string{key, str})
			key = ""
		}
	}
	return nil

}

func vals(s *goquery.Selection) *Values {
	n := s.Find("dt")
	ret := make(Values, len(n.Nodes))
	n.Each(func(i int, ns *goquery.Selection) {
		ret[i] = [2]string{strings.TrimSpace(ns.Text()), strings.TrimSpace(ns.Next().Text())}
	})
	return &ret
}

func scrape(url, prev string) (*Entry, string, string) {
	if url == "" {
		return nil, "", ""
	}
	doc, err := goquery.NewDocument(base + url)
	if err != nil {
		log.Fatal(err)
	}
	entry := &Entry{
		Name:    doc.Find(".archive").Text(),
		URL:     url,
		About:   vals(doc.Find("#about")),
		Contact: vals(doc.Find("#contact")),
	}
	next := doc.Find(".page-nav-next").Last().AttrOr("href", "")
	if next == prev {
		next = ""
	}
	return entry, next, url
}

func write(dir string, e []*Entry) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	os.Mkdir(dir, 0700)
	for i, v := range e {
		if err := enc.Encode(v); err != nil {
			return err
		}
		if err := ioutil.WriteFile(fmt.Sprintf("%s/%d.json", dir, i), buf.Bytes(), 0644); err != nil {
			return err
		}
		i++
		buf.Reset()
	}
	return nil
}

func load(dir string) ([]*Entry, error) {
	var i int
	ret := make([]*Entry, 0, 1000)
	for {
		f, err := os.Open(fmt.Sprintf("%s/%d.json", dir, i))
		if err != nil {
			break
		}
		e := &Entry{}
		dec := json.NewDecoder(f)
		err = dec.Decode(e)
		f.Close()
		if err != nil {
			return nil, err
		}
		i++
		ret = append(ret, e)
	}
	return ret, nil
}

func download(url string) []*Entry {
	ret := make([]*Entry, 0, 1000)
	for e, next, prev := scrape(url, ""); e != nil; e, next, prev = scrape(next, prev) {
		ret = append(ret, e)
	}
	return ret
}

func uniqs(e []*Entry) (map[string]int, map[string]int) {
	about, contact := make(map[string]int), make(map[string]int)
	for _, v := range e {
		for _, w := range *v.About {
			about[w[0]]++
		}
		for _, w := range *v.Contact {
			contact[w[0]]++
		}
	}
	return about, contact
}

func sample(e []*Entry, k string) {
	for _, v := range e {
		val, ok := v.About.get(k)
		if !ok {
			val, ok = v.Contact.get(k)
		}
		if ok {
			fmt.Println(val)
		}
	}
}

func main() {
	e, err := load("directory")
	if err != nil {
		log.Fatal(err)
	}
	rows := ToCSV(e)
	out, err := os.Create("directory/import.csv")
	defer out.Close()
	if err != nil {
		log.Fatal(err)
	}
	w := csv.NewWriter(out)
	log.Print(w.WriteAll(rows))
}
