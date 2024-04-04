package check

import (
	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/stretchr/testify/assert"
)

func (s *CheckESTestSuite) TestCheckTransformError() {

	// When check all transform
	monitoringData, err := s.check.CheckTransformError("_all", []string{})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check all indices with exclude
	monitoringData, err = s.check.CheckTransformError("_all", []string{"foo"})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check all indices and exclude al existing transform
	monitoringData, err = s.check.CheckTransformError("_all", []string{"test"})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_OK, monitoringData.Status())

	// When check transform that not exist
	monitoringData, err = s.check.CheckTransformError("foo", []string{})
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), monitoringData)
	assert.Equal(s.T(), nagiosPlugin.STATUS_UNKNOWN, monitoringData.Status())

}
