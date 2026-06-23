export namespace config {
	
	export class ProxyGroupPreview {
	    name: string;
	    type: string;
	    now: string;
	    all: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProxyGroupPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.now = source["now"];
	        this.all = source["all"];
	    }
	}
	export class ProxyPreviewResult {
	    groups: ProxyGroupPreview[];
	    proxyCount: number;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyPreviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groups = this.convertValues(source["groups"], ProxyGroupPreview);
	        this.proxyCount = source["proxyCount"];
	        this.mode = source["mode"];
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
	export class RuleEntry {
	    type: string;
	    payload: string;
	    proxy: string;
	
	    static createFrom(source: any = {}) {
	        return new RuleEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.payload = source["payload"];
	        this.proxy = source["proxy"];
	    }
	}

}

export namespace main {
	
	export class AppInfo {
	    name: string;
	    version: string;
	    websiteUrl: string;
	    helpDocsUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.websiteUrl = source["websiteUrl"];
	        this.helpDocsUrl = source["helpDocsUrl"];
	    }
	}
	export class ConfigEntry {
	    id: string;
	    name: string;
	    sourceUrl: string;
	    updatedAt: number;
	    isActive: boolean;
	    proxyCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ConfigEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sourceUrl = source["sourceUrl"];
	        this.updatedAt = source["updatedAt"];
	        this.isActive = source["isActive"];
	        this.proxyCount = source["proxyCount"];
	    }
	}
	export class RulesPageResult {
	    total: number;
	    rules: config.RuleEntry[];
	    offset: number;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new RulesPageResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.rules = this.convertValues(source["rules"], config.RuleEntry);
	        this.offset = source["offset"];
	        this.limit = source["limit"];
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
	export class Status {
	    running: boolean;
	    configPath: string;
	    controller: string;
	    mixedPort: number;
	    systemProxyEnabled: boolean;
	    systemProxyServer: string;
	    connected: boolean;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.configPath = source["configPath"];
	        this.controller = source["controller"];
	        this.mixedPort = source["mixedPort"];
	        this.systemProxyEnabled = source["systemProxyEnabled"];
	        this.systemProxyServer = source["systemProxyServer"];
	        this.connected = source["connected"];
	        this.version = source["version"];
	    }
	}
	export class TrafficStats {
	    upload: number;
	    download: number;
	    upTotal: number;
	    downTotal: number;
	
	    static createFrom(source: any = {}) {
	        return new TrafficStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.upload = source["upload"];
	        this.download = source["download"];
	        this.upTotal = source["upTotal"];
	        this.downTotal = source["downTotal"];
	    }
	}
	export class UpdateCheckResult {
	    hasUpdate: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.message = source["message"];
	    }
	}

}

export namespace profile {
	
	export class Profile {
	    id: string;
	    name: string;
	    filename: string;
	    sourceUrl?: string;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.filename = source["filename"];
	        this.sourceUrl = source["sourceUrl"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

export namespace settings {
	
	export class Settings {
	    autoStartCore: boolean;
	    autoSystemProxy: boolean;
	    systemProxyEnabled: boolean;
	    mixedPort: number;
	    tunEnabled: boolean;
	    allowLan: boolean;
	    logLevel: string;
	    subscriptionUserAgent: string;
	    activeProfileId: string;
	    proxySelections: Record<string, any>;
	    startMinimized: boolean;
	    launchAtLogin: boolean;
	    themeMode: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoStartCore = source["autoStartCore"];
	        this.autoSystemProxy = source["autoSystemProxy"];
	        this.systemProxyEnabled = source["systemProxyEnabled"];
	        this.mixedPort = source["mixedPort"];
	        this.tunEnabled = source["tunEnabled"];
	        this.allowLan = source["allowLan"];
	        this.logLevel = source["logLevel"];
	        this.subscriptionUserAgent = source["subscriptionUserAgent"];
	        this.activeProfileId = source["activeProfileId"];
	        this.proxySelections = source["proxySelections"];
	        this.startMinimized = source["startMinimized"];
	        this.launchAtLogin = source["launchAtLogin"];
	        this.themeMode = source["themeMode"];
	    }
	}

}

export namespace subscription {
	
	export class Item {
	    id: string;
	    name: string;
	    url: string;
	    updatedAt: number;
	    trafficUsed?: string;
	    trafficTotal?: string;
	    expireAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.updatedAt = source["updatedAt"];
	        this.trafficUsed = source["trafficUsed"];
	        this.trafficTotal = source["trafficTotal"];
	        this.expireAt = source["expireAt"];
	    }
	}

}

