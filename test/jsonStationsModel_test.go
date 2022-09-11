package noxon

import (
	"testing"

	"git.privatehive.de/bjoern/noxon-server/pkg/noxon"
	"github.com/stretchr/testify/assert"
)

func requestDataRecursive(t *testing.T, parentId *string, model noxon.JsonModel) {

	for i := 0; i < model.Count(parentId); i++ {
		item, id := model.Data(parentId, i)
		assert.NotEmpty(t, id)
		assert.NotEmpty(t, item)
		itemTwo, idTwo := model.Data(&id, -1)
		assert.NotEmpty(t, idTwo)
		assert.NotEmpty(t, itemTwo)
		assert.Equal(t, id, idTwo)
		requestDataRecursive(t, &id, model)
	}
}

func TestOperationHealthCheck(t *testing.T) {

	model := noxon.NewJsonStationsModel()
	requestDataRecursive(t, nil, model)
}
