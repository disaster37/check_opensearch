package check

import (
	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/stretchr/testify/assert"
)

func (s *CheckESTestSuite) TestCheckIndiceLocked() {

	// When check all indices
	monitoringData, err := s.check.CheckIndiceLocked("_all")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_CRITICAL, monitoringData.Status())

	// When check only one indice
	monitoringData, err = s.check.CheckIndiceLocked("test")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check indice that not exist
	monitoringData, err = s.check.CheckIndiceLocked("foo")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_UNKNOWN, monitoringData.Status())

	// When indice is locked and only one indice
	monitoringData, err = s.check.CheckIndiceLocked("lock")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_CRITICAL, monitoringData.Status())
}
