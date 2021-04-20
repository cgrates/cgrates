/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ColCfg = "config"
)

func (ms *MongoStorage) GeneralJsonCfg() (cfg *config.GeneralJsonCfg, err error) {
	cfg = new(config.GeneralJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.GeneralJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.GeneralJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) RPCConnJsonCfg() (cfg config.RPCConnsJson, err error) {
	cfg = make(config.RPCConnsJson)
	r := &struct {
		Section string
		Cfg     config.RPCConnsJson
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.RPCConnsJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) CacheJsonCfg() (cfg *config.CacheJsonCfg, err error) {
	cfg = new(config.CacheJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.CacheJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.CacheJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ListenJsonCfg() (cfg *config.ListenJsonCfg, err error) {
	cfg = new(config.ListenJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.ListenJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.ListenJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) HttpJsonCfg() (cfg *config.HTTPJsonCfg, err error) {
	cfg = new(config.HTTPJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.HTTPJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.HTTPJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) DbJsonCfg(section string) (cfg *config.DbJsonCfg, err error) {
	cfg = new(config.DbJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.DbJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": section})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) FilterSJsonCfg() (cfg *config.FilterSJsonCfg, err error) {
	cfg = new(config.FilterSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.FilterSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.FilterSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) CdrsJsonCfg() (cfg *config.CdrsJsonCfg, err error) {
	cfg = new(config.CdrsJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.CdrsJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.CDRsJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ERsJsonCfg() (cfg *config.ERsJsonCfg, err error) {
	cfg = new(config.ERsJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.ERsJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.ERsJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) EEsJsonCfg() (cfg *config.EEsJsonCfg, err error) {
	cfg = new(config.EEsJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.EEsJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.EEsJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) SessionSJsonCfg() (cfg *config.SessionSJsonCfg, err error) {
	cfg = new(config.SessionSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.SessionSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.SessionSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) FreeswitchAgentJsonCfg() (cfg *config.FreeswitchAgentJsonCfg, err error) {
	cfg = new(config.FreeswitchAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.FreeswitchAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.FreeSWITCHAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) KamAgentJsonCfg() (cfg *config.KamAgentJsonCfg, err error) {
	cfg = new(config.KamAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.KamAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.KamailioAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) AsteriskAgentJsonCfg() (cfg *config.AsteriskAgentJsonCfg, err error) {
	cfg = new(config.AsteriskAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.AsteriskAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.AsteriskAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) DiameterAgentJsonCfg() (cfg *config.DiameterAgentJsonCfg, err error) {
	cfg = new(config.DiameterAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.DiameterAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.DiameterAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) RadiusAgentJsonCfg() (cfg *config.RadiusAgentJsonCfg, err error) {
	cfg = new(config.RadiusAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.RadiusAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.RadiusAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) HttpAgentJsonCfg() (cfg *[]*config.HttpAgentJsonCfg, err error) {
	cfg = new([]*config.HttpAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *[]*config.HttpAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.HTTPAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) DNSAgentJsonCfg() (cfg *config.DNSAgentJsonCfg, err error) {
	cfg = new(config.DNSAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.DNSAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.DNSAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) AttributeServJsonCfg() (cfg *config.AttributeSJsonCfg, err error) {
	cfg = new(config.AttributeSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.AttributeSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.AttributeSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ChargerServJsonCfg() (cfg *config.ChargerSJsonCfg, err error) {
	cfg = new(config.ChargerSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.ChargerSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.ChargerSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ResourceSJsonCfg() (cfg *config.ResourceSJsonCfg, err error) {
	cfg = new(config.ResourceSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.ResourceSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.ResourceSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) StatSJsonCfg() (cfg *config.StatServJsonCfg, err error) {
	cfg = new(config.StatServJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.StatServJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.StatSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ThresholdSJsonCfg() (cfg *config.ThresholdSJsonCfg, err error) {
	cfg = new(config.ThresholdSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.ThresholdSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.ThresholdSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) RouteSJsonCfg() (cfg *config.RouteSJsonCfg, err error) {
	cfg = new(config.RouteSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.RouteSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.RouteSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) LoaderJsonCfg() (cfg []*config.LoaderJsonCfg, err error) {
	cfg = make([]*config.LoaderJsonCfg, 0)
	r := &struct {
		Section string
		Cfg     []*config.LoaderJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.LoaderSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) MailerJsonCfg() (cfg *config.MailerJsonCfg, err error) {
	cfg = new(config.MailerJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.MailerJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.MailerJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) SureTaxJsonCfg() (cfg *config.SureTaxJsonCfg, err error) {
	cfg = new(config.SureTaxJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.SureTaxJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.SureTaxJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) DispatcherSJsonCfg() (cfg *config.DispatcherSJsonCfg, err error) {
	cfg = new(config.DispatcherSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.DispatcherSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.DispatcherSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) RegistrarCJsonCfgs() (cfg *config.RegistrarCJsonCfgs, err error) {
	cfg = new(config.RegistrarCJsonCfgs)
	r := &struct {
		Section string
		Cfg     *config.RegistrarCJsonCfgs
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.RegistrarCJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) LoaderCfgJson() (cfg *config.LoaderCfgJson, err error) {
	cfg = new(config.LoaderCfgJson)
	r := &struct {
		Section string
		Cfg     *config.LoaderCfgJson
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.LoaderJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) MigratorCfgJson() (cfg *config.MigratorCfgJson, err error) {
	cfg = new(config.MigratorCfgJson)
	r := &struct {
		Section string
		Cfg     *config.MigratorCfgJson
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.MigratorJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) TlsCfgJson() (cfg *config.TlsJsonCfg, err error) {
	cfg = new(config.TlsJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.TlsJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.TlsJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) AnalyzerCfgJson() (cfg *config.AnalyzerSJsonCfg, err error) {
	cfg = new(config.AnalyzerSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.AnalyzerSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.AnalyzerSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) AdminSCfgJson() (cfg *config.AdminSJsonCfg, err error) {
	cfg = new(config.AdminSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.AdminSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.AdminSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) RateCfgJson() (cfg *config.RateSJsonCfg, err error) {
	cfg = new(config.RateSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.RateSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.RateSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) SIPAgentJsonCfg() (cfg *config.SIPAgentJsonCfg, err error) {
	cfg = new(config.SIPAgentJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.SIPAgentJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.SIPAgentJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) TemplateSJsonCfg() (cfg config.FcTemplatesJsonCfg, err error) {
	cfg = make(config.FcTemplatesJsonCfg)
	r := &struct {
		Section string
		Cfg     config.FcTemplatesJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.TemplatesJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ConfigSJsonCfg() (cfg *config.ConfigSCfgJson, err error) {
	cfg = new(config.ConfigSCfgJson)
	r := &struct {
		Section string
		Cfg     *config.ConfigSCfgJson
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.ConfigSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ApiBanCfgJson() (cfg *config.APIBanJsonCfg, err error) {
	cfg = new(config.APIBanJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.APIBanJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.APIBanJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) CoreSJSON() (cfg *config.CoreSJsonCfg, err error) {
	cfg = new(config.CoreSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.CoreSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.CoreSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) ActionSCfgJson() (cfg *config.ActionSJsonCfg, err error) {
	cfg = new(config.ActionSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.ActionSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.ActionSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}
func (ms *MongoStorage) AccountSCfgJson() (cfg *config.AccountSJsonCfg, err error) {
	cfg = new(config.AccountSJsonCfg)
	r := &struct {
		Section string
		Cfg     *config.AccountSJsonCfg
	}{Cfg: cfg}
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": config.AccountSJSON})
		if err := cur.Decode(r); err != nil {
			cfg = nil
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetSection(ctx *context.Context, section string, jsn interface{}) (err error) {
	return ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColCfg).UpdateOne(sctx, bson.M{"section": section},
			bson.M{"$set": &struct {
				Section string
				Cfg     interface{}
			}{Section: section, Cfg: jsn}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}
