package middlewares

import (
	"fmt"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/pkg/dbx"
	"github.com/getfider/fider/app/pkg/oauth"
	"github.com/getfider/fider/app/pkg/web"
	"github.com/getfider/fider/app/storage/postgres"
)

//Setup current context with some services
func Setup(db *dbx.Database) web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			tx, err := db.Begin()
			if err != nil {
				return err
			}

			c.SetActiveTransaction(tx)
			c.SetServices(&app.Services{
				Tenants: &postgres.TenantStorage{Trx: tx},
			})

			defer func() {
				if r := recover(); r != nil {
					var err error
					switch r := r.(type) {
					case error:
						err = r
					default:
						err = fmt.Errorf("%v", r)
					}
					c.Error(err)
				}
			}()

			err = next(c)
			if err == nil {
				return tx.Commit()
			}
			tx.Rollback()
			return err
		}
	}
}

//AddServices adds services to current context
func AddServices() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c web.Context) error {
			tx := c.ActiveTransaction()
			tenant := c.Tenant()
			services := c.Services()
			services.OAuth = &oauth.HTTPService{}
			services.Ideas = &postgres.IdeaStorage{Trx: tx, Tenant: tenant}
			services.Users = &postgres.UserStorage{Trx: tx}
			services.Tenants = &postgres.TenantStorage{Trx: tx}

			return next(c)
		}
	}
}