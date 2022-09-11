package noxon

type NullStationsModel struct {
}

func NewNullStationsModel() NullStationsModel {

	return NullStationsModel{}
}

func (m NullStationsModel) Data(parentId *string, index int) (Item, string) {

	return ItemDir{}, ""
}

func (m NullStationsModel) Count(parentId *string) int {

	return 0
}
