package data_models

import (
  "github.com/nttdots/go-dots/dots_server/db"
  "github.com/nttdots/go-dots/dots_server/models"
  "github.com/nttdots/go-dots/dots_server/db_models/data"
  log "github.com/sirupsen/logrus"
)

type Client struct {
  Id       int64
  Customer *models.Customer
  Cuid     string
  Cdid     *string
}

func NewClient(customer *models.Customer, cuid string, cdid *string) Client {
  if customer == nil {
    panic("customer must not be nil.")
  }

  return Client {
    Customer: customer,
    Cuid: cuid,
    Cdid: cdid,
  }
}

func (client *Client) Save(tx *db.Tx) error {

  c := data_db_models.Client{}
  c.Id         = client.Id
  c.CustomerId = client.Customer.Id
  c.Cuid       = client.Cuid
  c.Cdid       = db.AsNullString(client.Cdid)

  if c.Id == 0 {
    _, err := tx.Session.Insert(&c)
    if err != nil {
      log.WithError(err).Errorf("Insert() failed.")
      return err
    } else {
      client.Id = c.Id
      return nil
    }
  } else {
    _, err := tx.Session.ID(c.Id).Update(&c)
    if err != nil {
      log.WithError(err).Errorf("Update() failed.")
      return err
    } else {
      return nil
    }
  }
}

func FindClientByCuid(tx *db.Tx, customer *models.Customer, cuid string) (*Client, error) {
  c := data_db_models.Client{}
  has, err := tx.Session.Where("customer_id=? AND cuid=?", customer.Id, cuid).Get(&c)
  if err != nil {
    log.WithError(err).Error("Get() failed.")
    return nil, err
  }

  if !has {
    return nil, nil
  } else {
    return &Client{
      Id:       c.Id,
      Customer: customer,
      Cuid:     c.Cuid,
      Cdid:     db.AsStringPointer(c.Cdid),
    }, nil
  }
}

func CheckExistDotsClient(tx *db.Tx, cuid string) (bool, error) {
  c := data_db_models.Client{}
  has, err := tx.Session.Where("cuid=?", cuid).Get(&c)
  if err != nil {
    log.WithError(err).Error("Get() failed.")
    return false, err
  }
  if !has {
    return false, nil
  }
  return true, nil
}

func DeleteClientByCuid(tx *db.Tx, customer *models.Customer, cuid string) (bool, error) {
  dbClient, err := FindClientByCuid(tx, customer, cuid)
  if err != nil{
    return false, err
  }

  // not found client
  if dbClient == nil{
    return false, nil
  }

  // Delete alias table data
  _, err = tx.Session.Delete(&data_db_models.Alias{ClientId: dbClient.Id})
  if err != nil {
		log.Errorf("Failed to delete alias: %s", err)
		return false, err
  }

  // Delete acl table data
  _, err = tx.Session.Delete(&data_db_models.ACL{ClientId: dbClient.Id})
  if err != nil {
     log.Errorf("Failed to delete acl: %s", err)
     return false, err
  }

  // Delete client table data
  affected, err := tx.Session.Delete(&data_db_models.Client{Id: dbClient.Id})
  if err != nil {
     log.Errorf("Failed to delete client: %s", err)
     return false, err
  }

  return 0 < affected, nil
}
