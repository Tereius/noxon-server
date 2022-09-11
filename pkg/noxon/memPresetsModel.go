package noxon

type MemPresetsModel struct {
	presets map[string]string
}

func NewMemPresetsModel() MemPresetsModel {

	return MemPresetsModel{
		presets: map[string]string{},
	}
}

func (m MemPresetsModel) WritePreset(presetKey string, stationId string) error {

	m.presets[presetKey] = stationId
	return nil
}

func (m MemPresetsModel) GetPreset(presetKey string) string {

	return m.presets[presetKey]
}
