
include "apitypes.thrift"

service RunWorkService {
    bool connected()
    void connect()
    void interrupt()
    void interruptDelay()
    void switchGas(1:i8 valve)
    void searchProducts(1:string comportName)
    void runDeviceWork()
    void runLuaScript(1:string filename)
    void sendDeviceCommand(1:apitypes.RequestDeviceCommand reqCmd)
}

service WorkDialogService {
    void setConfigParamValues (1:list<apitypes.ConfigParamValue> configParamValues)
    list<apitypes.ConfigParamValue> getConfigParamValues ()
    void IgnoreError()
    void selectWorks(1:list<bool> works)
    void selectWork(1:i32 workIndex)
}

service CurrentFileService {
    void requestChart()
    void renameChart(1:string oldName, 2:string newName)
    void addNewProducts(1:i8 productsCount)
    void deleteProducts(1:list<i64> productIDs)
    list<apitypes.DeviceParam> listDeviceParams()
    void runEdit()
    void openFile(1:string filename)
    list<apitypes.WorkLogRecord> listWorkLogRecords()

    void exportChart(1:string filename, 2:apitypes.TimeUnixMillis timeFrom, 3:apitypes.TimeUnixMillis timeTo, 4:string chart)
}

service ProductParamService {
    void setValue(1:string key, 2:i64 productID, 3:string value)
    string getValue(1:string key, 2:i64 productID)
}

service ProductService {
   void setProductSerial(1:i64 productID, 2:i64 serial)
   void setProductsComport(1:list<i64> productIDs, 2:string comport)
   void setProductAddr(1:i64 productID, 2:i16 addr)
   void setProductActive(1:i64 productID, 2:bool active)
   void setProductParamSeries(1:apitypes.ProductParamSeries productParam)
   apitypes.ProductParamSeries getProductParamSeries(1:i64 productID, 2:i16 paramAddr)
   void deleteChartPoints(1:apitypes.DeleteChartPointsRequest r)

   void setNetAddr(1:i64 productID)
}

service FilesService {
   void createNewParty(1:i8 productsCount)
   apitypes.Party getCurrentParty()
   void setCurrentParty(1:i64 partyID)
   apitypes.Party getParty(1:i64 partyID)
   list<apitypes.PartyInfo> listParties(1:i64 filterSerilal)
   void deleteFile(1:i64 partyID)
   void saveFile(1:i64 partyID, 2:string filename)
   void copyFile(1:i64 partyID)
}

service FileService {
   apitypes.PartyProductsValues getProductsValues(1:i64 partyID, 2:i64 filterSerial)
}

service NotifyGuiService {
    oneway void open(1:i64 hWnd)
    oneway void close()
}

service AppConfigService {
    void editConfig()
    list<string> listDevices()
    apitypes.DeviceInfo currentDeviceInfo()

    list<apitypes.ConfigParamValue> getParamValues()
    void setParamValue(1:string key, 2:string value)
    string getParamValue(1:string key)
}

service HelperService {
    string FormatWrite32BCD(1:string s)
    string FormatWrite32FloatBE(1:string s)
    string FormatWrite32FloatLE(1:string s)
}

service TemperatureDeviceService {
    void start()
    void stop()
    void setup(1:double temperature)
    void setDestination(1:double temperature)
    double getTemperature()
    void coolingOn()
    void coolingOff()
}

service CoefficientsService {
    void writeAll(1:list<apitypes.ProductCoefficientValue> xs)
    void readAll()
    void setActive(1:i32 n, 2:bool active)
    list<apitypes.ProductCoefficientValue> getCurrentPartyCoefficients()
}

service JournalService {
    list<string> listDays()
    list<apitypes.JournalEntry> listEntriesOfDay(1:string day)
    list<i64> listEntriesIDsOfDay(1:string day)
    apitypes.JournalEntry getEntryByID(1:i64 entryID)
    void deleteDays(1:list<string> days)
}

service AppInfoService {
    apitypes.BuildInfo buildInfo()
}

