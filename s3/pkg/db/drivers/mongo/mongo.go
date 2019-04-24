// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
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

package mongo

import (
	"github.com/globalsign/mgo"
)

var adap = &adapter{}
var DataBaseName = "metadatastore"
var BucketMD = "metadatabucket"
//var StorageClassCol = "StorageClasses"

func Init(host string) *adapter {
	//fmt.Println("edps:", deps)
	session, err := mgo.Dial(host)
	if err != nil {
		panic(err)
	}
	//defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	adap.s = session
	adap.userID = "unknown"

	/*//Init default storage class definition.
	err = adap.initDefaultStorageClasses()
	if err != nil {
		panic(err)
	}*/

	return adap
}

func Exit() {
	adap.s.Close()
}

type adapter struct {
	s      *mgo.Session
	userID string
}

//Init default storage classes, include STANDARD, STANDARD_IA, GLACIER.
/*func (ad *adapter) initDefaultStorageClasses() error {
	c := ad.s.DB(DataBaseName).C(StorageClassCol)
	scs := []StorageClassDef{}
	err := c.Find(bson.M{}).All(&scs)
	if err != nil {
		log.Logf("Get storage class definition failed:%v.\n", err)
		return err
	}

	//Add the default storage classes to database.
	if len(scs) == 0 {
		standard := StorageClassDef{
			Name:"STNADARD",
			SupportedBackendTypes: []string{
				constants.BackendTypeAws,
				constants.BackendTypeAzure,
				constants.BackendFusionStorage,
				constants.BackendTypeCeph,
				constants.BackendTypeObs,
				constants.BackendTypeGcp,
				constants.BackendTypeIBMCos,
			},
			NextStorageClasses:[]string{"STANDARD_IA", "GLACIER"}}
		err = c.Insert(standard)
		if err != nil {
			log.Logf("Insert standard storage class to database failed:%v.\n", err)
			return err
		}
		standardIA := StorageClassDef{
			Name:"STNADARD_IA",
			SupportedBackendTypes: []string{
				constants.BackendTypeAws,
				constants.BackendTypeAzure,
				constants.BackendTypeObs,
				constants.BackendTypeGcp,
				constants.BackendTypeIBMCos,
			},
			NextStorageClasses:[]string{"GLACIER"}}
		err = c.Insert(standardIA)
		if err != nil {
			log.Logf("Insert standard_ia storage class to database failed:%v.\n", err)
			return err
		}
		glacier := StorageClassDef{
			Name:"GLACIER",
			SupportedBackendTypes: []string{
				constants.BackendTypeAws,
				constants.BackendTypeAzure,
				constants.BackendTypeObs,
				constants.BackendTypeGcp,
				constants.BackendTypeIBMCos,
			}}
		err = c.Insert(glacier)
		if err != nil {
			log.Logf("Insert glacier storage class to database failed:%v.\n", err)
			return err
		}
	}

	return nil
}

func (ad *adapter) CheckStorageClassValidation(sc *string) (error, string) {
	//if *sc == "" {
	//	return nil, "STANDARD"
	//}

	c := ad.s.DB(DataBaseName).C(StorageClassCol)
	var scDef StorageClassDef
	err := c.Find(bson.M{"name":sc}).One(&scDef)
	if err == mgo.ErrNotFound {
		log.Logf("Storage class [%s] has not defined.\n", sc)
		return nil, ""
	}
	if err != nil {
		log.Logf("Try to find storage class [%s] in database failed:%v.\n", err)
		return err, ""
	}

	return nil, scDef.Name
}

//Get supported backend types of a specific storage class.
func (ad *adapter) GetSupportedBackendTypes(sc *string) (error, []string) {
	c := ad.s.DB(DataBaseName).C(StorageClassCol)
	var scDef StorageClassDef
	//err := c.Find(bson.M{"supportedbackendtypes":sc}).All(scDef)
	err := c.Find(bson.M{"name":sc}).One(scDef)
	if err != nil {
		log.Logf("Get supported backend types for %s failed, err:%v.\n", sc, err)
		return err, nil
	}

	return nil, scDef.SupportedBackendTypes
}*/