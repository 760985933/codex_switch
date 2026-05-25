export namespace main {
	
	export class AppConfig {
	    listenHost: string;
	    listenPort: number;
	    deepseekBaseURL: string;
	    apiKey: string;
	    defaultModel: string;
	    requestTimeoutMs: number;
	    maxRetries: number;
	    enableAutoStart: boolean;
	    minimizeToTray: boolean;
	    logRetentionDays: number;
	    compactMode: boolean;
	    mappings: Record<string, string>;
	    headers: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.listenHost = source["listenHost"];
	        this.listenPort = source["listenPort"];
	        this.deepseekBaseURL = source["deepseekBaseURL"];
	        this.apiKey = source["apiKey"];
	        this.defaultModel = source["defaultModel"];
	        this.requestTimeoutMs = source["requestTimeoutMs"];
	        this.maxRetries = source["maxRetries"];
	        this.enableAutoStart = source["enableAutoStart"];
	        this.minimizeToTray = source["minimizeToTray"];
	        this.logRetentionDays = source["logRetentionDays"];
	        this.compactMode = source["compactMode"];
	        this.mappings = source["mappings"];
	        this.headers = source["headers"];
	    }
	}
	export class BridgeStatusPayload {
	    status: string;
	    listenAddress: string;
	    startedAt: string;
	    uptimeSeconds: number;
	    lastError: string;
	    requestCount: number;
	
	    static createFrom(source: any = {}) {
	        return new BridgeStatusPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.listenAddress = source["listenAddress"];
	        this.startedAt = source["startedAt"];
	        this.uptimeSeconds = source["uptimeSeconds"];
	        this.lastError = source["lastError"];
	        this.requestCount = source["requestCount"];
	    }
	}
	export class HealthCheckItem {
	    name: string;
	    ok: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthCheckItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ok = source["ok"];
	        this.message = source["message"];
	    }
	}
	export class HealthCheckResult {
	    ok: boolean;
	    checks: HealthCheckItem[];
	
	    static createFrom(source: any = {}) {
	        return new HealthCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.checks = this.convertValues(source["checks"], HealthCheckItem);
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
	export class LogEntry {
	    id: string;
	    level: string;
	    timestamp: string;
	    source: string;
	    message: string;
	    requestId?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.timestamp = source["timestamp"];
	        this.source = source["source"];
	        this.message = source["message"];
	        this.requestId = source["requestId"];
	    }
	}
	export class OverviewSnapshot {
	    config: AppConfig;
	    status: BridgeStatusPayload;
	    recentLogs: LogEntry[];
	    quickTips: string[];
	    defaults: Record<string, string>;
	    features: Record<string, boolean>;
	
	    static createFrom(source: any = {}) {
	        return new OverviewSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], AppConfig);
	        this.status = this.convertValues(source["status"], BridgeStatusPayload);
	        this.recentLogs = this.convertValues(source["recentLogs"], LogEntry);
	        this.quickTips = source["quickTips"];
	        this.defaults = source["defaults"];
	        this.features = source["features"];
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

}

