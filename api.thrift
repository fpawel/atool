
include "apitypes.thrift"

service ProductsService {
   void createNewParty(1:i8 productsCount, 2:string note)
   void setPartyNote(1:i64 partyID, 2:string note)
   apitypes.Party getParty(1:i64 partyID)
   apitypes.Party getCurrentParty()
   list<apitypes.PartyInfo> listParties()
   void setCurrentParty(1:i64 partyID)
   void addNewProducts(1:i8 productsCount)
   void deleteProducts(1:list<i64> productIDs)
   void setProductsComport(1:list<i64> productIDs, 2:string comport)
   void setProductsDevice(1:list<i64> productIDs, 2:string device)
   void setProduct(1:apitypes.Product product)

   list<string> listDevices()

   string getAppConfig()
   void setAppConfig(1:string appConfig)

   list<apitypes.YearMonth> listYearMonths()
}