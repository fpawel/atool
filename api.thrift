
include "apitypes.thrift"

service ProductsService {
   oneway void createNewParty()
   apitypes.Party getParty(1:i64 partyID)
   void addNewProduct()
   void deleteProduct(1:i64 productID)
   void setProduct(1:apitypes.Product product)

   list<string> getComports()
   void setComports(1:list<string> comports)

   string getAppConfig()
   void setAppConfig(1:string appConfig)

   list<apitypes.YearMonth> listYearMonths()
}