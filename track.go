/*
 * Created by dengshiwei on 2022/06/06.
 * Copyright 2015－2022 Sensors Data Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sensorsanalytics

import (
	"fmt"
	"github.com/sensorsdata/sa-sdk-go/structs"
	"github.com/sensorsdata/sa-sdk-go/utils"
	"os"
	"runtime"
)

const (
	SDK_VERSION = "2.1.0"
	LIB_NAME    = "Golang"
)

func TrackEvent(sa *SensorsAnalytics, etype, event, distinctId, originId string, properties map[string]interface{}, isLoginId bool) error {
	eventTime := utils.NowMs()
	if properties == nil {
		properties = map[string]interface{}{}
	}
	if et := extractUserTime(properties); et > 0 {
		eventTime = et
	}

	data := structs.EventData{
		Type:          etype,
		Time:          eventTime,
		DistinctId:    distinctId,
		Properties:    properties,
		LibProperties: getLibProperties(),
	}

	if sa.ProjectName != "" {
		data.Project = sa.ProjectName
	}

	if etype == TRACK || etype == TRACK_SIGNUP {
		data.Event = event
		properties["$lib"] = LIB_NAME
		properties["$lib_version"] = SDK_VERSION
	}

	if etype == TRACK_SIGNUP {
		data.OriginId = originId
	}

	if sa.TimeFree {
		data.TimeFree = true
	}

	if isLoginId {
		properties["$is_login_id"] = true
	}

	err := data.NormalizeData()
	if err != nil {
		return err
	}

	return sa.C.Send(data)
}

func ItemTrack(sa *SensorsAnalytics, trackType string, itemType string, itemId string, properties map[string]interface{}) error {
	libProperties := getLibProperties()
	time := utils.NowMs()
	if properties == nil {
		properties = map[string]interface{}{}
	}

	itemData := structs.Item{
		Type:          trackType,
		ItemId:        itemId,
		Time:          time,
		ItemType:      itemType,
		Properties:    properties,
		LibProperties: libProperties,
	}

	err := itemData.NormalizeItem()
	if err != nil {
		return err
	}

	return sa.C.ItemSend(itemData)
}

func TrackEventID3(sa *SensorsAnalytics, identity Identity, etype, event string, properties map[string]interface{}) error {
	eventTime := utils.NowMs()
	if properties == nil {
		properties = map[string]interface{}{}
	}

	if et := extractUserTime(properties); et > 0 {
		eventTime = et
	}

	data := structs.EventData{
		Type:          etype,
		Time:          eventTime,
		Identities:    identity.Identities,
		Properties:    properties,
		LibProperties: getLibProperties(),
	}

	err := data.CheckIdentities()
	if err != nil {
		return err
	}

	// 添加 distinct_id
	distinctId := identity.Identities[LOGIN_ID]
	if len(distinctId) <= 0 {
		for _, v := range identity.Identities {
			distinctId = v
		}
	}
	data.DistinctId = distinctId

	if sa.ProjectName != "" {
		data.Project = sa.ProjectName
	}

	if etype == TRACK || etype == BIND || etype == UNBIND {
		data.Event = event
		properties["$lib"] = LIB_NAME
		properties["$lib_version"] = SDK_VERSION
	}

	if sa.TimeFree {
		data.TimeFree = true
	}

	err = data.NormalizeData()
	if err != nil {
		return err
	}

	return sa.C.Send(data)
}

func getLibProperties() structs.LibProperties {
	lp := structs.LibProperties{}
	lp.Lib = LIB_NAME
	lp.LibVersion = SDK_VERSION
	lp.LibMethod = "code"
	if pc, file, line, ok := runtime.Caller(3); ok { //3 means sdk's caller
		f := runtime.FuncForPC(pc)
		lp.LibDetail = fmt.Sprintf("##%s##%s##%d", f.Name(), file, line)
	}

	return lp
}

func extractUserTime(p map[string]interface{}) int64 {
	if t, ok := p["$time"]; ok {
		v, ok := t.(int64)
		if !ok {
			fmt.Fprintln(os.Stderr, "It's not ok for type string")
			return 0
		}
		delete(p, "$time")

		return v
	}

	return 0
}
