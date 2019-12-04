
include "apitypes.thrift"

service ProductsService {
   void createNewParty(1:i8 productsCount)
   apitypes.Party getParty(1:i64 partyID)
   apitypes.Party getCurrentParty()
   void requestCurrentPartyChart()
   list<apitypes.PartyInfo> listParties()
   void setCurrentParty(1:i64 partyID)
   void addNewProducts(1:i8 productsCount)
   void deleteProducts(1:list<i64> productIDs)
   void setProductsComport(1:list<i64> productIDs, 2:string comport)
   void setProductsDevice(1:list<i64> productIDs, 2:string device)
   void setProductAddr(1:i64 productID, 2:i16 addr)
   void setProductActive(1:i64 productID, 2:bool active)
   void setProductParam(1:apitypes.ProductParam productParam)
   apitypes.ProductParam getProductParam(1:i64 productID, 2:i16 paramAddr)
   list<string> listDevices()
   void EditConfig()

   oneway void openGuiClient(1:i64 hWnd)
   oneway void closeGuiClient()

   bool Connected()
   void Connect()
   void Disconnect()
}