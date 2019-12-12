
include "apitypes.thrift"

service HardwareConnectionService {
    bool connected()
    void connect()
    void disconnect()
}

service CurrentFileService {
    void requestCurrentPartyChart()
    void setPartyName(1:string name)

    void addNewProducts(1:i8 productsCount)
    void deleteProducts(1:list<i64> productIDs)
}

service ProductService {
   void setProductSerial(1:i64 productID, 2:i64 serial)
   void setProductsComport(1:list<i64> productIDs, 2:string comport)
   void setProductsDevice(1:list<i64> productIDs, 2:string device)
   void setProductAddr(1:i64 productID, 2:i16 addr)
   void setProductActive(1:i64 productID, 2:bool active)
   void setProductParam(1:apitypes.ProductParam productParam)
   apitypes.ProductParam getProductParam(1:i64 productID, 2:i16 paramAddr)
   void deleteChartPoints(1:apitypes.DeleteChartPointsRequest r)
}

service MainService {
   void createNewParty(1:i8 productsCount, 2:string name)
   apitypes.Party getCurrentParty()
   void setCurrentParty(1:i64 partyID)
   apitypes.Party getParty(1:i64 partyID)
   list<apitypes.PartyInfo> listParties()
   list<string> listDevices()
   list<i32> listParamAddresses()
   void EditConfig()
   oneway void openGuiClient(1:i64 hWnd)
   oneway void closeGuiClient()
}