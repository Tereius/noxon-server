package noxon

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"

	b64 "encoding/base64"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

const macObfuscate = "a6703ded78821be5"
const loginEndpoint = "/setupapp/fs/asp/BrowseXML/loginXML.asp"
const searchEndpoint = "/setupapp/fs/asp/BrowseXML/Search.asp"
const addPresetEndpoint = "/Favorites/AddPreset.aspx"
const getPresetEndpoint = "/Favorites/GetPreset.aspx"
const playbackEndpoint = "/playback"
const healthEndpoint = "/health"
const statusEndpoint = "/status"

type ListOfItems struct {
	XMLName   xml.Name `xml:"ListOfItems"`
	ItemCount int      `xml:"ItemCount"`
	Items     []Item
}

type Item interface {
	build(c *gin.Context, id string) Item
	toString() string
}

type ItemMessage struct {
	XMLName  xml.Name `xml:"Item"`
	ItemType string   `xml:"ItemType"`
	Message  string   `xml:"Message"`
}

type ItemDir struct {
	XMLName  xml.Name `xml:"Item"`
	ItemType string   `xml:"ItemType"`
	Title    string   `xml:"Title"`
	UrlDir   string   `xml:"UrlDir"`
}

type ItemStation struct {
	XMLName            xml.Name `xml:"Item"`
	ItemType           string   `xml:"ItemType"`
	StationId          string   `xml:"StationId"` // Limit 32 Byte?
	StationName        string   `xml:"StationName"`
	StationUrl         string   `xml:"StationUrl"`
	StationDescription string   `xml:"StationDesc"`
	StationFormat      string   `xml:"StationFormat"`    // Public
	StationBandWidth   string   `xml:"StationBandWidth"` // 128
	StationMime        string   `xml:"StationMime"`      // MP3
}

type Redirect struct {
	Location string
}

func (e *Redirect) Error() string {
	return fmt.Sprintf("Redirect to %s", e.Location)
}

func (i ItemMessage) build(c *gin.Context, id string) Item {
	i.ItemType = "Message"
	return i
}

func (i ItemDir) build(c *gin.Context, id string) Item {
	i.ItemType = "Dir"
	i.UrlDir = getBasePath(c) + loginEndpoint + "?gofile=" + b64.URLEncoding.EncodeToString([]byte(id))
	return i
}

func (i ItemStation) build(c *gin.Context, id string) Item {
	i.ItemType = "Station"
	i.StationId = b64.URLEncoding.EncodeToString([]byte(id))
	i.StationUrl = buildPlaybackUrl(c, id)
	return i
}

func buildPlaybackUrl(c *gin.Context, stationId string) string {
	return getBasePath(c) + playbackEndpoint + "?mac=" + c.Query("mac") + "&stationId=" + b64.URLEncoding.EncodeToString([]byte(stationId))
}

func (i ItemMessage) toString() string { return "Message" }

func (i ItemDir) toString() string { return "Dir" }

func (i ItemStation) toString() string { return "Station" }

type StationsModel interface {
	Data(parentId *string, index int) (Item, string)
	Count(parentId *string) int
}

type PresetModel interface {
	WritePreset(presetKey string, stationId string) error
	GetPreset(presetKey string) string
}

type NoxonServer struct {
	engine      *gin.Engine
	settings    NoxonServerSettings
	presetMutex sync.Mutex
}

type encryptedToken struct {
	XMLName xml.Name `xml:"EncryptedToken"`
	Token   string   `xml:",chardata"`
}

func NewNoxonServer(settings NoxonServerSettings) *NoxonServer {

	return &NoxonServer{
		engine:      gin.New(),
		settings:    settings,
		presetMutex: sync.Mutex{},
	}
}

func getBasePath(c *gin.Context) string {

	if c != nil && c.Request != nil && c.Request.URL != nil {
		return fmt.Sprintf("%s://%s", "http", c.Request.Host)
	}
	return ""
}

type DeviceInfo struct {
	Mac             string
	FirmwareVersion string
	HardwareVersion string
	Vendor          string
	Language        string
}

func extractDeviceInfo(c *gin.Context) (ret DeviceInfo) {

	if c != nil {
		ret.Mac = c.Query("mac")
		ret.FirmwareVersion = c.Query("fver")
		ret.HardwareVersion = c.Query("hw")
		ret.Vendor = c.Query("ven")
		ret.Language = c.Query("dlang")
	}
	return ret
}

func min(one int, two int) int {
	if one < two {
		return one
	}
	return two
}

func (n *NoxonServer) CollectFromModel(c *gin.Context, parent *string, start int, end int) (ret []Item) {

	count := n.settings.StationsModel.Count(parent)
	trueEnd := min(end+1, count)
	for i := start; i < trueEnd; i++ {
		item, id := n.settings.StationsModel.Data(parent, i)
		if len(id) == 0 {
			log.Warn("Got invalid Item id")
		}
		ret = append(ret, item.build(c, id))
	}
	return ret
}

func (n *NoxonServer) authMiddleware(c *gin.Context) {

	device := extractDeviceInfo(c)
	log := log.WithField("device", device)
	accessGranted := false
	if c.Request.URL.Path == loginEndpoint && len(c.Query("token")) > 0 {
		// Login query always accepted
		accessGranted = true
	} else if c.Request.URL.Path == healthEndpoint {
		// Health query always accepted
		accessGranted = true
	} else {
		mac := device.Mac
		if len(mac) > 0 {
			for _, blackListedMac := range n.settings.Blacklist {
				if blackListedMac == mac || blackListedMac == "*" {
					accessGranted = false
					break
				}
			}
			// Whitelist overrules Blacklist
			for _, whiteListedMac := range n.settings.Whitelist {
				if whiteListedMac == mac || whiteListedMac == "*" {
					accessGranted = true
					break
				}
			}
		}
	}
	if accessGranted {
		c.Next()
	} else {
		log.Infof("Access denied")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func (n *NoxonServer) handleLoginEndpoint(c *gin.Context) {

	log := log.WithField("device", extractDeviceInfo(c))
	firstItem, _ := strconv.Atoi(c.DefaultQuery("startitems", "1"))
	lastItem, _ := strconv.Atoi(c.DefaultQuery("enditems", fmt.Sprintf("%d", firstItem+99)))
	// Those crazy noxon people start count with 1 - we correct that
	firstItem--
	lastItem--
	if token := c.Query("token"); len(token) > 0 {
		log.Debug("Login request")
		c.XML(200, encryptedToken{Token: macObfuscate})
	} else if gofile := c.Query("gofile"); gofile == "" {
		// Request the root menu (No pagination is happening here)
		log.Debug("Root menu request")
		rootItemsCount := n.settings.StationsModel.Count(nil)
		if rootItemsCount > 0 {
			ItemList := ListOfItems{
				ItemCount: rootItemsCount,
				Items:     n.CollectFromModel(c, nil, firstItem, rootItemsCount-1),
			}
			writeXmlResponse(c, ItemList)
		} else {
			writeMessageResponse(c, "No stations found")
		}
	} else if gofile := c.Query("gofile"); gofile != "" {
		// Request a submenu
		itemId, err := b64.URLEncoding.DecodeString(gofile)
		itemIdString := string(itemId)
		log.Debugf("Submenu request for parent itemId %s (%d - %d)", itemIdString, firstItem, lastItem)
		if err != nil {
			log.Errorf("Could not decode itemId: %s", err.Error())
			c.AbortWithStatus(http.StatusBadRequest)
		} else {
			itemIdString := itemIdString
			ItemList := ListOfItems{
				ItemCount: n.settings.StationsModel.Count(&itemIdString),
				Items:     n.CollectFromModel(c, &itemIdString, firstItem, lastItem),
			}
			writeXmlResponse(c, ItemList)
		}
	} else {
		c.AbortWithStatus(404)
	}
}

func (n *NoxonServer) handleSearchEndpoint(c *gin.Context) {

	log := log.WithField("device", extractDeviceInfo(c))
	if searchId := c.Query("Search"); searchId != "" {
		itemId, err := b64.URLEncoding.DecodeString(searchId)
		itemIdString := string(itemId)
		log.Debugf("Search request for itemId %s", itemIdString)
		if err != nil {
			log.Errorf("Could not decode itemId: %s", err.Error())
			c.AbortWithStatus(http.StatusBadRequest)
		} else {
			stationItem, stationItemId := n.settings.StationsModel.Data(&itemIdString, -1)
			if len(stationItemId) > 0 {
				ItemList := ListOfItems{
					ItemCount: -1,
					Items:     []Item{stationItem.build(c, itemIdString)},
				}
				writeXmlResponse(c, ItemList)
			} else {
				log.Errorf("A non existing item (id: %s) was requested", itemIdString)
				c.AbortWithStatus(404)
			}
		}
	} else {
		c.AbortWithStatus(404)
	}
}

func (n *NoxonServer) handleAddPresetEndpoint(c *gin.Context) {

	device := extractDeviceInfo(c)
	log := log.WithField("device", device)
	mutex.Lock()
	currentStationId := stationTracker[device.Mac]
	mutex.Unlock()
	if presetIndex := c.Query("id"); presetIndex != "" && len(currentStationId) > 0 {

		log.Infof("Saving stationId %s to preset %s", currentStationId, presetIndex)

		n.presetMutex.Lock()
		err := n.settings.PresetsModel.WritePreset(device.Mac+"-"+presetIndex, currentStationId)
		n.presetMutex.Unlock()

		if err != nil {
			writeMessageResponse(c, "Preset could not be saved")
		} else {
			writeMessageResponse(c, "Preset saved")
		}
	} else {
		writeMessageResponse(c, "Preset not created - select a station and try again")
	}
}

func (n *NoxonServer) handleGetPresetEndpoint(c *gin.Context) {

	device := extractDeviceInfo(c)
	//log := log.WithField("device", device)
	if presetIndex := c.Query("id"); presetIndex != "" {

		n.presetMutex.Lock()
		stationId := n.settings.PresetsModel.GetPreset(device.Mac + "-" + presetIndex)
		n.presetMutex.Unlock()

		stationItem, stationItemId := n.settings.StationsModel.Data(&stationId, -1)
		if len(stationItemId) > 0 {
			ItemList := ListOfItems{
				ItemCount: -1,
				Items:     []Item{stationItem.build(c, stationItemId)},
			}
			writeXmlResponse(c, ItemList)
		} else {
			writeMessageResponse(c, "Preset not set")
		}
	} else {
		writeMessageResponse(c, "Preset not set")
	}
}

var mutex = sync.Mutex{}
var stationUrls = map[string]string{}    // Maps stationId's to stream url's. We need this because the stream url (from the model) might be redirected (and then differs from the model)
var stationTracker = map[string]string{} // Maps Device macs with the currently playedback stationId
var proxyHistory = map[string]string{}   // History of proxy reqests

func (n *NoxonServer) handlePlaybackEndpoint(c *gin.Context) {

	device := extractDeviceInfo(c)
	log := log.WithField("device", device)
	if stationIdQuery := c.Query("stationId"); stationIdQuery != "" {
		//redirectCounter, _ := strconv.Atoi(c.DefaultQuery("counter", 0))
		stationId, err := b64.URLEncoding.DecodeString(stationIdQuery)
		stationIdString := string(stationId)
		log.Debugf("Playback request for stationId %s", stationIdString)
		if err != nil {
			log.Errorf("Could not decode stationId: %s", err.Error())
			c.AbortWithStatus(http.StatusBadRequest)
		} else {
			mutex.Lock()
			streamUrl := stationUrls[stationIdString]
			mutex.Unlock()
			if len(streamUrl) == 0 {
				// request the url from the model
				stationItem, stationItemId := n.settings.StationsModel.Data(&stationIdString, -1)
				if station, ok := stationItem.(ItemStation); ok && len(stationItemId) > 0 {
					streamUrl = station.StationUrl
					mutex.Lock()
					stationUrls[stationIdString] = streamUrl
					mutex.Unlock()
				} else {
					log.Errorf("A non existing item (id: %s) was requested", stationItemId)
					c.AbortWithStatus(404)
					return
				}
			}

			if remote, err := url.Parse(streamUrl); err != nil || len(streamUrl) == 0 {
				log.Errorf("Could not parse streamUrl: %s (%v)", streamUrl, err)
				c.AbortWithStatus(http.StatusInternalServerError)
			} else {
				proxy := httputil.NewSingleHostReverseProxy(remote)
				proxy.Director = func(req *http.Request) {
					req.Header = c.Request.Header
					req.Host = remote.Host
					req.URL = remote
				}
				proxy.ModifyResponse = func(r *http.Response) error {
					if r.StatusCode == http.StatusMovedPermanently || r.StatusCode == http.StatusFound || r.StatusCode == http.StatusPermanentRedirect || r.StatusCode == http.StatusTemporaryRedirect {
						log.Info("Got redirect request")
						if newStreamUrl := r.Header["Location"]; len(newStreamUrl) > 0 {
							return &Redirect{
								Location: newStreamUrl[0],
							}
						} else {
							return fmt.Errorf("missing new redirect location")
						}
					}
					return nil
				}
				proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, err error) {

					if redirect, ok := err.(*Redirect); ok {
						log.Infof("Forwarding redirect to new location %s", redirect.Location)
						mutex.Lock()
						stationUrls[stationIdString] = redirect.Location
						mutex.Unlock()
						http.Redirect(rw, r, buildPlaybackUrl(c, stationIdString), http.StatusFound)
					} else {
						log.Errorf("Proxy error: %s", err.Error())
						rw.WriteHeader(http.StatusBadGateway)
					}
				}

				mutex.Lock()
				// Device starts playback
				stationTracker[device.Mac] = stationIdString
				proxyHistory[remote.String()] = device.Mac
				mutex.Unlock()

				log.Infof("Starting proxy for target url: %s", remote.String())
				proxy.ServeHTTP(c.Writer, c.Request)

				mutex.Lock()
				// Device stops playback
				delete(stationTracker, device.Mac)
				mutex.Unlock()
			}
		}
	} else {
		c.AbortWithStatus(404)
	}
}

func (n *NoxonServer) handleStatusEndpoint(c *gin.Context) {

	mutex.Lock()
	activePlayback := "<ul>"
	// Device stops playback
	for _, stationId := range stationTracker {
		activePlayback += fmt.Sprintf(`<li>StationId: %s - StationUrl: <a href="%s">%s</a></li>`, stationId, stationUrls[stationId], stationUrls[stationId])
	}
	activePlayback += "</ul>"
	mutex.Unlock()

	mutex.Lock()
	playbackHistory := "<ul>"
	// Device stops playback
	for url := range proxyHistory {
		playbackHistory += fmt.Sprintf(`<li><a href="%s">%s</a></li>`, url, url)
	}
	playbackHistory += "</ul>"
	mutex.Unlock()

	data := []byte(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>Noxon status</title>
	</head>
	<body>
		<h2>Active playback</h2>
		` + activePlayback + `
		<h2>Playback history</h2>
		` + playbackHistory + `
	</body>
	</html>
	`)
	c.Data(200, "text/html", data)
}

func (n *NoxonServer) handleHealthEndpoint(c *gin.Context) {

	data := []byte(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>Noxon health check</title>
	</head>
	<body>
		<h2>Ok</h2>
	</body>
	</html>
	`)
	c.Data(200, "text/html", data)
}

func (n *NoxonServer) handleRecovery(c *gin.Context, err any) {

	device := extractDeviceInfo(c)
	if err != nil && err == http.ErrAbortHandler {
		mutex.Lock()
		// Device stops playback
		delete(stationTracker, device.Mac)
		mutex.Unlock()
	}
}

func writeMessageResponse(c *gin.Context, message string) {

	ItemList := ListOfItems{
		ItemCount: -1,
		Items:     []Item{ItemMessage{Message: message}.build(c, "")},
	}
	writeXmlResponse(c, ItemList)
}

func writeXmlResponse(c *gin.Context, xmlStruct any) {

	if c != nil {
		data, err := xml.Marshal(xmlStruct)
		if err != nil {
			log.Warnf("Could not marshal xml %s", err.Error())
		} else {
			data = append([]byte(`<?xml version="1.0" encoding="iso-8859-1" standalone="yes"?>`), data...)
			c.Data(200, "text/html", data)
		}
	} else {
		log.Error("Gin context not available")
	}
}

func (n *NoxonServer) StartAndServe() {

	log.Infof("Starting noxon server")
	n.engine.Use(ginlogrus.Logger(log.New()))
	n.engine.Use(gin.CustomRecoveryWithWriter(nil, n.handleRecovery))
	n.engine.Use(n.authMiddleware)
	n.engine.GET(loginEndpoint, n.handleLoginEndpoint)
	n.engine.GET(searchEndpoint, n.handleSearchEndpoint)
	n.engine.GET(playbackEndpoint, n.handlePlaybackEndpoint)
	n.engine.GET(addPresetEndpoint, n.handleAddPresetEndpoint)
	n.engine.GET(getPresetEndpoint, n.handleGetPresetEndpoint)
	n.engine.GET(healthEndpoint, n.handleHealthEndpoint)
	n.engine.GET(statusEndpoint, n.handleStatusEndpoint)
	n.engine.Run("0.0.0.0:80")
}
