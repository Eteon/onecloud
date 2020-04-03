// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"context"
	"strings"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/sqlchemy"

	"yunion.io/x/onecloud/pkg/apis/monitor"
	api "yunion.io/x/onecloud/pkg/apis/monitor"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

var (
	SuggestSysAlertManager *SSuggestSysAlertManager
)

type SSuggestSysAlertManager struct {
	db.SVirtualResourceBaseManager
	db.SEnabledResourceBaseManager
}

func init() {
	SuggestSysAlertManager = &SSuggestSysAlertManager{
		SVirtualResourceBaseManager: db.NewVirtualResourceBaseManager(
			&SSuggestSysAlert{},
			"suggestsysalert_tbl",
			"suggestsysalert",
			"suggestsysalerts",
		),
	}
	SuggestSysAlertManager.SetVirtualObject(SuggestSysAlertManager)
}

type SSuggestSysAlert struct {
	db.SVirtualResourceBase
	db.SEnabledResourceBase

	//监控规则对应的json对象
	RuleName      string               `list:"user" update:"user"`
	MonitorConfig jsonutils.JSONObject `list:"user" create:"required" update:"user"`
	//监控规则type：Rule Type
	Type    string               `width:"256" charset:"ascii" list:"user" update:"user"`
	ResMeta jsonutils.JSONObject `list:"user" update:"user"`
	Problem jsonutils.JSONObject `list:"user" update:"user"`
	Suggest string               `width:"256"  list:"user" update:"user"`
	Action  string               `width:"256" charset:"ascii" list:"user" update:"user"`
	ResId   string               `width:"256" charset:"ascii" list:"user" update:"user"`
}

func NewSuggestSysAlertManager(dt interface{}, keyword, keywordPlural string) *SSuggestSysAlertManager {
	man := &SSuggestSysAlertManager{
		SVirtualResourceBaseManager: db.NewVirtualResourceBaseManager(
			dt,
			"sugalart_tbl",
			keyword,
			keywordPlural,
		),
	}
	man.SetVirtualObject(man)
	return man
}

func (manager *SSuggestSysAlertManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query monitor.SuggestSysAlertListInput) (*sqlchemy.SQuery, error) {
	var err error
	q, err = manager.SVirtualResourceBaseManager.ListItemFilter(ctx, q, userCred, query.VirtualResourceListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVirtualResourceBaseManager.ListItemFilter")
	}
	q, err = manager.SEnabledResourceBaseManager.ListItemFilter(ctx, q, userCred, query.EnabledResourceBaseListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SEnabledResourceBaseManager.ListItemFilter")
	}
	if len(query.Type) > 0 {
		q = q.Equals("type", query.Type)
	}
	if len(query.ResId) > 0 {
		q = q.Equals("res_id", query.ResId)
	}
	return q, nil
}

func (manager *SSuggestSysAlertManager) CustomizeFilterList(ctx context.Context, q *sqlchemy.SQuery, userCred mcclient.TokenCredential, query jsonutils.JSONObject) (*db.CustomizeListFilters, error) {
	filters := db.NewCustomizeListFilters()
	input := new(monitor.SuggestSysAlertListInput)
	if err := query.Unmarshal(input); err != nil {
		return nil, err
	}
	wrapF := func(key string, inputValue []string) func(object jsonutils.JSONObject) (bool, error) {
		return func(data jsonutils.JSONObject) (bool, error) {
			id, err := data.GetString("id")
			if err != nil {
				return false, err
			}
			obj, err := manager.GetAlert(id)
			if err != nil {
				return false, err
			}
			value, err := obj.ResMeta.GetString(key)
			if err != nil {
				return false, errors.Wrapf(err, "SSuggestSysAlert's ResMeta get %s error,id is %s", key, obj.GetId())
			}
			return strings.Contains(value, strings.Join(inputValue, ",")), nil
		}
	}
	if input.CloudEnv != "" {
		filters.Append(wrapF("cloud_env", []string{input.CloudEnv}))
	}
	if len(input.Brands) > 0 {
		filters.Append(wrapF("brand", input.Brands))
	}
	if len(input.Providers) > 0 {
		filters.Append(wrapF("provider", input.Providers))
	}
	if input.Tenant != "" {
		filters.Append(wrapF("tenant", []string{input.Project}))
	}
	if input.Account != "" {
		filters.Append(wrapF("account", []string{input.Cloudaccount}))
	}

	return filters, nil
}

func (manager *SSuggestSysAlertManager) GetAlert(id string) (*SSuggestSysAlert, error) {
	obj, err := manager.FetchById(id)
	if err != nil {
		return nil, err
	}
	return obj.(*SSuggestSysAlert), nil
}

func (man *SSuggestSysAlertManager) OrderByExtraFields(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	input monitor.SuggestSysAlertListInput,
) (*sqlchemy.SQuery, error) {
	var err error
	q, err = man.SVirtualResourceBaseManager.OrderByExtraFields(ctx, q, userCred, input.VirtualResourceListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVirtualResourceBaseManager.OrderByExtraFields")
	}
	return q, nil
}

func (man *SSuggestSysAlertManager) ValidateCreateData(
	ctx context.Context, userCred mcclient.TokenCredential,
	ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject,
	data monitor.SuggestSysAlertCreateInput) (monitor.SuggestSysAlertCreateInput, error) {
	//rule 查询到资源信息后没有将资源id，进行转换
	if len(data.ResID) == 0 {
		return data, httperrors.NewInputParameterError("not found res_id %q", data.ResID)
	}
	if len(data.Type) == 0 {
		return data, httperrors.NewInputParameterError("not found type %q", data.Type)
	}
	return data, nil
}

func (man *SSuggestSysAlertManager) FetchCustomizeColumns(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	objs []interface{},
	fields stringutils2.SSortedStrings,
	isList bool,
) []monitor.SuggestSysAlertDetails {
	rows := make([]monitor.SuggestSysAlertDetails, len(objs))
	virtRows := man.SVirtualResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, objs, fields, isList)
	for i := range rows {
		rows[i] = monitor.SuggestSysAlertDetails{
			VirtualResourceDetails: virtRows[i],
		}
		rows[i] = objs[i].(*SSuggestSysAlert).getMoreDetails(rows[i])
	}
	return rows
}

func (self *SSuggestSysAlert) getMoreDetails(out monitor.SuggestSysAlertDetails) monitor.SuggestSysAlertDetails {
	err := self.ResMeta.Unmarshal(&out)
	if err != nil {
		log.Errorln("SSuggestSysAlert getMoreDetails's error:", err)
	}
	suggestSysSettingMap, _ := SuggestSysRuleManager.FetchSuggestSysAlartSettings(self.Type)
	out.RuleName = suggestSysSettingMap[self.Type].Name
	out.ResType = GetSuggestSysRuleDrivers()[self.Type].GetResourceType()

	return out
}

func (manager *SSuggestSysAlertManager) QueryDistinctExtraField(q *sqlchemy.SQuery, field string) (*sqlchemy.SQuery, error) {
	var err error
	q, err = manager.SVirtualResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}
	return q, httperrors.ErrNotFound
}

func (alert *SSuggestSysAlert) ValidateUpdateData(
	ctx context.Context, userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	data monitor.SuggestSysAlertUpdateInput) (monitor.SuggestSysAlertUpdateInput, error) {
	//rule 查询到资源信息后没有将资源id，进行转换
	if len(data.ResID) == 0 {
		return data, httperrors.NewInputParameterError("not found res_id ")
	}
	if len(data.Type) == 0 {
		return data, httperrors.NewInputParameterError("not found type ")
	}
	var err error
	data.VirtualResourceBaseUpdateInput, err = alert.SVirtualResourceBase.ValidateUpdateData(ctx, userCred, query,
		data.VirtualResourceBaseUpdateInput)
	if err != nil {
		return data, errors.Wrap(err, "SVirtualResourceBase.ValidateUpdateData")
	}
	return data, nil
}

func (self *SSuggestSysAlert) GetExtraDetails(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	isList bool,
) (monitor.SuggestSysAlertDetails, error) {
	return monitor.SuggestSysAlertDetails{}, nil
}

func (self *SSuggestSysAlert) PostDelete(ctx context.Context, userCred mcclient.TokenCredential) {

}

func (self *SSuggestSysAlert) CustomizeDelete(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) error {
	return self.StartDeleteTask(ctx, userCred)
}

func (self *SSuggestSysAlert) Delete(ctx context.Context, userCred mcclient.TokenCredential) error {
	log.Infof("SSuggestSysAlert delete do nothing")
	return nil
}

func (self *SSuggestSysAlert) RealDelete(ctx context.Context, userCred mcclient.TokenCredential) error {
	return self.SVirtualResourceBase.Delete(ctx, userCred)
}

func (self *SSuggestSysAlert) StartDeleteTask(
	ctx context.Context, userCred mcclient.TokenCredential) error {
	params := jsonutils.NewDict()
	self.SetStatus(userCred, api.EIP_UNUSED_START_DELETE, "")
	return GetSuggestSysRuleDrivers()[self.Type].StartDeleteTask(ctx, userCred, self, params)
}
