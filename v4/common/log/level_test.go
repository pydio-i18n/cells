package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetDynamicDebugLevels(t *testing.T) {
	Convey("Test DynamicDebugLevels in/out", t, func() {
		SetDynamicDebugLevels(false, true, "service.name")
		SetDynamicDebugLevels(false, true, "service.name")
		So(dynamicDebug, ShouldHaveLength, 1)
		SetDynamicDebugLevels(false, true, "service.name2")
		So(dynamicDebug, ShouldHaveLength, 2)
		So(ddRegexp, ShouldHaveLength, 2)
		SetDynamicDebugLevels(false, false, "service.name")
		So(dynamicDebug, ShouldHaveLength, 1)
		SetDynamicDebugLevels(true, false)
		So(dynamicDebug, ShouldHaveLength, 0)
	})
}
