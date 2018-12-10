/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017,-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"configcenter/src/common"
	"configcenter/src/common/errors"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/common/universalsql/mongo"
	"configcenter/src/source_controller/coreservice/core"
	"configcenter/src/storage/dal"
)

type modelAttribute struct {
	model   *modelManager
	dbProxy dal.RDB
}

func (m *modelAttribute) CreateModelAttributes(ctx core.ContextParams, objID string, inputParam metadata.CreateModelAttributes) (dataResult *metadata.CreateManyDataResult, err error) {

	dataResult = &metadata.CreateManyDataResult{}

	addExceptionFunc := func(idx int64, err errors.CCErrorCoder, attr *metadata.Attribute) {
		dataResult.CreateManyInfoResult.Exceptions = append(dataResult.CreateManyInfoResult.Exceptions, metadata.ExceptionResult{
			OriginIndex: idx,
			Message:     err.Error(),
			Code:        int64(err.GetCode()),
			Data:        attr,
		})
	}

	for attrIdx, attr := range inputParam.Attributes {

		_, exists, err := m.IsExists(ctx, attr.PropertyID)

		if nil != err {
			addExceptionFunc(int64(attrIdx), err.(errors.CCErrorCoder), &attr)
			continue
		}

		if exists {
			dataResult.CreateManyInfoResult.Repeated = append(dataResult.CreateManyInfoResult.Repeated, metadata.RepeatedDataResult{
				OriginIndex: int64(attrIdx),
				Data:        mapstr.NewFromStruct(attr, "field"),
			})
			continue
		}
		id, err := m.Save(ctx, attr)
		if nil != err {
			addExceptionFunc(int64(attrIdx), err.(errors.CCErrorCoder), &attr)
			continue
		}

		dataResult.CreateManyInfoResult.Created = append(dataResult.CreateManyInfoResult.Created, metadata.CreatedDataResult{
			OriginIndex: int64(attrIdx),
			ID:          id,
		})

	}
	err = nil // unused ,reserved
	return dataResult, err
}

func (m *modelAttribute) SetModelAttributes(ctx core.ContextParams, objID string, inputParam metadata.SetModelAttributes) (dataResult *metadata.SetDataResult, err error) {

	dataResult = &metadata.SetDataResult{}

	addExceptionFunc := func(idx int64, err errors.CCErrorCoder, attr *metadata.Attribute) {
		dataResult.Exceptions = append(dataResult.Exceptions, metadata.ExceptionResult{
			OriginIndex: idx,
			Message:     err.Error(),
			Code:        int64(err.GetCode()),
			Data:        attr,
		})
	}

	for attrIdx, attr := range inputParam.Attributes {

		existsAttr, exists, err := m.IsExists(ctx, attr.PropertyID)

		if nil != err {
			addExceptionFunc(int64(attrIdx), err.(errors.CCErrorCoder), &attr)
			continue
		}

		if exists {
			cond := mongo.NewCondition()
			cond.Element(&mongo.Eq{Key: metadata.AttributeFieldSupplierAccount, Val: ctx.SupplierAccount})
			cond.Element(&mongo.Eq{Key: metadata.AttributeFieldID, Val: existsAttr.ID})

			_, err := m.Update(ctx, mapstr.NewFromStruct(attr, "field"), cond)
			if nil != err {
				addExceptionFunc(int64(attrIdx), err.(errors.CCErrorCoder), &attr)
				continue
			}
			dataResult.Updated = append(dataResult.Updated, metadata.UpdatedDataResult{
				OriginIndex: int64(attrIdx),
				ID:          uint64(existsAttr.ID),
			})
			continue
		}
		id, err := m.Save(ctx, attr)
		if nil != err {
			addExceptionFunc(int64(attrIdx), err.(errors.CCErrorCoder), &attr)
			continue
		}

		dataResult.Created = append(dataResult.Created, metadata.CreatedDataResult{
			OriginIndex: int64(attrIdx),
			ID:          id,
		})

	}
	err = nil // unused ,reserved
	return dataResult, err
}
func (m *modelAttribute) UpdateModelAttributes(ctx core.ContextParams, objID string, inputParam metadata.UpdateOption) (*metadata.UpdatedCount, error) {

	cond, err := mongo.NewConditionFromMapStr(inputParam.Condition)
	if nil != err {
		return &metadata.UpdatedCount{}, err
	}

	cnt, err := m.Update(ctx, inputParam.Data, cond)
	if nil != err {
		return &metadata.UpdatedCount{}, err
	}

	return &metadata.UpdatedCount{Count: cnt}, nil
}
func (m *modelAttribute) DeleteModelAttributes(ctx core.ContextParams, objID string, inputParam metadata.DeleteOption) (*metadata.DeletedCount, error) {

	cond, err := mongo.NewConditionFromMapStr(inputParam.Condition)
	if nil != err {
		return &metadata.DeletedCount{}, err
	}

	cnt, err := m.dbProxy.Table(common.BKTableNameObjDes).Find(cond.ToMapStr()).Count(ctx)
	if nil != err {
		return &metadata.DeletedCount{}, err
	}

	if 0 == cnt {
		return &metadata.DeletedCount{}, nil
	}

	err = m.dbProxy.Table(common.BKTableNameObjAttDes).Delete(ctx, cond.ToMapStr())
	return &metadata.DeletedCount{Count: cnt}, err
}

func (m *modelAttribute) SearchModelAttributes(ctx core.ContextParams, objID string, inputParam metadata.QueryCondition) (*metadata.QueryResult, error) {

	cond, err := mongo.NewConditionFromMapStr(inputParam.Condition)
	if nil != err {
		return &metadata.QueryResult{}, err
	}

	attrResult := []metadata.Attribute{}
	err = m.dbProxy.Table(common.BKTableNameObjAttDes).Find(cond.ToMapStr()).All(ctx, &attrResult)
	if nil != err {
		return &metadata.QueryResult{}, err
	}

	dataResult := &metadata.QueryResult{Count: uint64(len(attrResult))}
	for _, attrItem := range attrResult {
		dataResult.Info = append(dataResult.Info, mapstr.NewFromStruct(attrItem, "field"))
	}
	return dataResult, nil
}
