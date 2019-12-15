
include "apitypes.thrift"

service HardwareConnectionService {
    bool connected()
    void connect()
    void disconnect()
    void command(1:i16 cmd; 2:string s)
    void switchGas(1:i8 deviceType, 2:string comport, 3:i8 valve)
}

service CurrentFileService {
    void requestChart()
    void setName(1:string name)
    void renameChart(1:string oldName, 2:string newName)
    void addNewProducts(1:i8 productsCount)
    void deleteProducts(1:list<i64> productIDs)
    list<i32> listParamAddresses()
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

service FilesService {
   void createNewParty(1:i8 productsCount, 2:string name)
   apitypes.Party getCurrentParty()
   void setCurrentParty(1:i64 partyID)
   apitypes.Party getParty(1:i64 partyID)
   list<apitypes.PartyInfo> listParties()
}

service NotifyGuiService {
    oneway void open(1:i64 hWnd)
    oneway void close()
}

service AppConfigService {
    void editConfig()
    list<string> listDevices()
    apitypes.AppConfig get()
}

service HelperService {
    string FormatWrite32BCD(1:string s)
    string FormatWrite32FloatBE(1:string s)
    string FormatWrite32FloatLE(1:string s)
}