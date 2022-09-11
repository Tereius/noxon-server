package noxon

type NoxonServerSettings struct {
	PresetsModel  PresetModel
	StationsModel StationsModel
	Whitelist     []string
	Blacklist     []string
}

func NewDefaultNoxonServerSettings() NoxonServerSettings {

	return NoxonServerSettings{
		PresetsModel:  NewMemPresetsModel(),
		StationsModel: NewNullStationsModel(),
		Whitelist:     []string{},
		Blacklist:     []string{},
	}
}

func (s NoxonServerSettings) WithPresetsModel(model PresetModel) NoxonServerSettings {

	s.PresetsModel = model
	return s
}

func (s NoxonServerSettings) WithStationsModel(model StationsModel) NoxonServerSettings {

	s.StationsModel = model
	return s
}

func (s NoxonServerSettings) WithBlacklist(list []string) NoxonServerSettings {

	s.Blacklist = list
	return s
}

func (s NoxonServerSettings) WithWhitelist(list []string) NoxonServerSettings {

	s.Whitelist = list
	return s
}
