package noxon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

type Entry struct {
	Id                 string
	DirName            string   `json:"dirName"`
	StationName        string   `json:"stationName"`
	StationDescription string   `json:"stationDescription"`
	StationUrl         string   `json:"stationUrl"`
	Children           []*Entry `json:"children"`
}

func (e *Entry) isDir() bool {
	return len(e.DirName) > 0
}

func (e *Entry) isStation() bool {
	return len(e.StationName) > 0
}

type JsonModel struct {
	data []*Entry
}

// Only used for debugging
func NewRandomJsonModel() JsonModel {
	data := []*Entry{}
	for i := 0; i < 20; i++ {
		data = append(data, &Entry{DirName: fmt.Sprintf("root (%d)", i)})
		for ii := 0; ii < 135; ii++ {
			data[i].Children = append(data[i].Children, &Entry{DirName: fmt.Sprintf("root/dir (%d, %d)", i, ii)})
			for iii := 0; iii < 20; iii++ {
				data[i].Children[ii].Children = append(data[i].Children[ii].Children, &Entry{StationName: fmt.Sprintf("root/dir/dir (%d, %d, %d) DLF", i, ii, iii), StationUrl: "https://st01.sslstream.dlf.de/dlf/01/128/mp3/stream.mp3?aggregator=web"})
			}
		}
	}
	jsonData, _ := json.Marshal(data)
	return NewJsonModelFromJson(jsonData)
}

func NewJsonModelFromJson(jsonData []byte) (ret JsonModel) {

	data := []*Entry{}
	json.Unmarshal(jsonData, &data)

	// index the model
	index := 0
	var indexer func(entries []*Entry)
	indexer = func(entries []*Entry) {
		for _, entry := range entries {
			entry.Id = fmt.Sprint(index)
			index++
			indexer(entry.Children)
		}
	}
	indexer(data)
	ret.data = data
	return ret
}

func NewJsonStationsModel() (ret JsonModel) {

	if jsonFile, err := os.Open("stations.json"); err != nil {
		log.Errorf("Could not read stations file: %s", err.Error())
	} else {
		defer jsonFile.Close()
		b, _ := ioutil.ReadAll(jsonFile)
		ret = NewJsonModelFromJson(b)
	}
	return ret
}

// TODO: Very unperformant recursive search - index the model in a map structure
func (m JsonModel) findEntry(id string) *Entry {

	var search func(entries []*Entry) *Entry
	search = func(entries []*Entry) *Entry {
		for _, entry := range entries {
			if entry != nil {
				if entry.Id == id {
					return entry
				}
				if child := search(entry.Children); child != nil {
					return child
				}
			}
		}
		return nil
	}
	return search(m.data)
}

func (m JsonModel) entryToItem(entry *Entry) (Item, string) {

	if entry != nil {
		if entry.isDir() {
			return ItemDir{
				Title: entry.DirName,
			}, entry.Id
		} else if entry.isStation() {
			return ItemStation{
				StationName:        entry.StationName,
				StationUrl:         entry.StationUrl,
				StationDescription: entry.StationDescription,
				StationMime:        "MP3",
			}, entry.Id
		}
	}
	return ItemDir{}, ""
}

func (m JsonModel) Data(parentId *string, index int) (Item, string) {

	if parentId == nil {
		// root
		return m.entryToItem(m.data[index])
	} else {
		if index >= 0 {
			// children
			if entry := m.findEntry(*parentId); entry != nil {
				return m.entryToItem(entry.Children[index])
			} else {
				log.Warnf("Could not find Item for parent '%s' with index %d", parentId, index)
			}
		} else {
			// the item with id
			return m.entryToItem(m.findEntry(*parentId))
		}
	}
	return ItemDir{}, ""
}

func (m JsonModel) Count(parentId *string) int {

	if parentId == nil {
		// root
		return len(m.data)
	} else {
		// children
		if entry := m.findEntry(*parentId); entry != nil {
			return len(entry.Children)
		}
	}
	return 0
}
