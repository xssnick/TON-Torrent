export namespace api {
	
	export class File {
	    Name: string;
	    Size: string;
	    Child: File[];
	    Path: string;
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Size = source["Size"];
	        this.Child = this.convertValues(source["Child"], File);
	        this.Path = source["Path"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NewProviderData {
	    Key: string;
	    MaxSpan: number;
	    PricePerMBDay: string;
	
	    static createFrom(source: any = {}) {
	        return new NewProviderData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.MaxSpan = source["MaxSpan"];
	        this.PricePerMBDay = source["PricePerMBDay"];
	    }
	}
	export class Peer {
	    IP: string;
	    ADNL: string;
	    Upload: string;
	    Download: string;
	
	    static createFrom(source: any = {}) {
	        return new Peer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IP = source["IP"];
	        this.ADNL = source["ADNL"];
	        this.Upload = source["Upload"];
	        this.Download = source["Download"];
	    }
	}
	export class PlainFile {
	    Path: string;
	    Name: string;
	    Size: string;
	    Downloaded: string;
	    Progress: number;
	    RawSize: number;
	
	    static createFrom(source: any = {}) {
	        return new PlainFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Name = source["Name"];
	        this.Size = source["Size"];
	        this.Downloaded = source["Downloaded"];
	        this.Progress = source["Progress"];
	        this.RawSize = source["RawSize"];
	    }
	}
	export class Provider {
	    Key: string;
	    LastProof: string;
	    PricePerDay: string;
	    Span: string;
	    Status: string;
	    Reason: string;
	    Peer: string;
	    Progress: number;
	    Data: NewProviderData;
	
	    static createFrom(source: any = {}) {
	        return new Provider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.LastProof = source["LastProof"];
	        this.PricePerDay = source["PricePerDay"];
	        this.Span = source["Span"];
	        this.Status = source["Status"];
	        this.Reason = source["Reason"];
	        this.Peer = source["Peer"];
	        this.Progress = source["Progress"];
	        this.Data = this.convertValues(source["Data"], NewProviderData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProviderContract {
	    Success: boolean;
	    Deployed: boolean;
	    Address: string;
	    Providers: Provider[];
	    Balance: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderContract(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.Deployed = source["Deployed"];
	        this.Address = source["Address"];
	        this.Providers = this.convertValues(source["Providers"], Provider);
	        this.Balance = source["Balance"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProviderRates {
	    Success: boolean;
	    Reason: string;
	    Provider: Provider;
	
	    static createFrom(source: any = {}) {
	        return new ProviderRates(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.Reason = source["Reason"];
	        this.Provider = this.convertValues(source["Provider"], Provider);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProviderStorageInfo {
	    Status: string;
	    Reason: string;
	    Downloaded: number;
	
	    static createFrom(source: any = {}) {
	        return new ProviderStorageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Status = source["Status"];
	        this.Reason = source["Reason"];
	        this.Downloaded = source["Downloaded"];
	    }
	}
	export class SpeedLimits {
	    Download: number;
	    Upload: number;
	
	    static createFrom(source: any = {}) {
	        return new SpeedLimits(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Download = source["Download"];
	        this.Upload = source["Upload"];
	    }
	}
	export class Torrent {
	    ID: string;
	    Name: string;
	    Size: string;
	    DownloadedSize: string;
	    Progress: number;
	    State: string;
	    Upload: string;
	    Download: string;
	    Path: string;
	    PeersNum: number;
	    Uploaded: string;
	    Ratio: string;
	
	    static createFrom(source: any = {}) {
	        return new Torrent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Size = source["Size"];
	        this.DownloadedSize = source["DownloadedSize"];
	        this.Progress = source["Progress"];
	        this.State = source["State"];
	        this.Upload = source["Upload"];
	        this.Download = source["Download"];
	        this.Path = source["Path"];
	        this.PeersNum = source["PeersNum"];
	        this.Uploaded = source["Uploaded"];
	        this.Ratio = source["Ratio"];
	    }
	}
	export class TorrentInfo {
	    Description: string;
	    Size: string;
	    Downloaded: string;
	    TimeLeft: string;
	    Progress: number;
	    State: string;
	    Upload: string;
	    Download: string;
	    Path: string;
	    Peers: number;
	    AddedAt: string;
	    Uploaded: string;
	    Ratio: string;
	
	    static createFrom(source: any = {}) {
	        return new TorrentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Description = source["Description"];
	        this.Size = source["Size"];
	        this.Downloaded = source["Downloaded"];
	        this.TimeLeft = source["TimeLeft"];
	        this.Progress = source["Progress"];
	        this.State = source["State"];
	        this.Upload = source["Upload"];
	        this.Download = source["Download"];
	        this.Path = source["Path"];
	        this.Peers = source["Peers"];
	        this.AddedAt = source["AddedAt"];
	        this.Uploaded = source["Uploaded"];
	        this.Ratio = source["Ratio"];
	    }
	}
	export class Transaction {
	    Body: string;
	    StateInit: string;
	    Address: string;
	    Amount: string;
	
	    static createFrom(source: any = {}) {
	        return new Transaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = source["Body"];
	        this.StateInit = source["StateInit"];
	        this.Address = source["Address"];
	        this.Amount = source["Amount"];
	    }
	}

}

export namespace main {
	
	export class Config {
	    DownloadsPath: string;
	    SeedMode: boolean;
	    ListenAddr: string;
	    Key: number[];
	    IsDarkTheme: boolean;
	    UseDaemon: boolean;
	    DaemonDBPath: string;
	    DaemonControlAddr: string;
	    PortsChecked: boolean;
	    NetworkConfigPath: string;
	    FetchIPOnStartup: boolean;
	    TunnelConfigPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DownloadsPath = source["DownloadsPath"];
	        this.SeedMode = source["SeedMode"];
	        this.ListenAddr = source["ListenAddr"];
	        this.Key = source["Key"];
	        this.IsDarkTheme = source["IsDarkTheme"];
	        this.UseDaemon = source["UseDaemon"];
	        this.DaemonDBPath = source["DaemonDBPath"];
	        this.DaemonControlAddr = source["DaemonControlAddr"];
	        this.PortsChecked = source["PortsChecked"];
	        this.NetworkConfigPath = source["NetworkConfigPath"];
	        this.FetchIPOnStartup = source["FetchIPOnStartup"];
	        this.TunnelConfigPath = source["TunnelConfigPath"];
	    }
	}
	export class TorrentAddResult {
	    Hash: string;
	    Err: string;
	
	    static createFrom(source: any = {}) {
	        return new TorrentAddResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hash = source["Hash"];
	        this.Err = source["Err"];
	    }
	}
	export class TorrentCreateResult {
	    Hash: string;
	    Err: string;
	
	    static createFrom(source: any = {}) {
	        return new TorrentCreateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hash = source["Hash"];
	        this.Err = source["Err"];
	    }
	}

}

