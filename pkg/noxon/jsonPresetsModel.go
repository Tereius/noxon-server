package noxon

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

type JsonPresetsModel struct {
	presets map[string]string
}

func NewJsonPresetsModel() JsonPresetsModel {

	model := JsonPresetsModel{
		presets: map[string]string{},
	}

	if dat, err := os.ReadFile("presets.json"); err != nil {
		log.Warnf("Could not read presets file - will create one if needed: %s", err.Error())
	} else if len(dat) > 0 {
		if err := json.Unmarshal(dat, &model.presets); err != nil {
			log.Errorf("Could not unmarshal presets: %s", err.Error())
		}
	}

	return model
}

func (m JsonPresetsModel) WritePreset(presetKey string, stationId string) error {

	m.presets[presetKey] = stationId
	if dat, err := json.Marshal(m.presets); err != nil {
		log.Errorf("Could not marshal presets: %s", err.Error())
		return err
	} else {
		if err := os.WriteFile("presets.json", dat, 0644); err != nil {
			log.Errorf("Could not write presets file: %s", err.Error())
			return err
		} else {
			log.Infof("Wrote presetKey '%s' with stationId '%s'", presetKey, stationId)
		}
	}
	return nil
}

func (m JsonPresetsModel) GetPreset(presetKey string) string {

	stationId := m.presets[presetKey]
	log.Infof("Read stationId '%s' from presetKey '%s'", stationId, presetKey)
	return stationId
}
