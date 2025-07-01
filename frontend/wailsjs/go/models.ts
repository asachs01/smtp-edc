export namespace services {

	export class CapabilityInfo {
	    pipelining: boolean;
	    startTLS: boolean;
	    auth: string[];
	    size: number;
	    eightBit: boolean;

	    static createFrom(source: any = {}) {
	        return new CapabilityInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pipelining = source["pipelining"];
	        this.startTLS = source["startTLS"];
	        this.auth = source["auth"];
	        this.size = source["size"];
	        this.eightBit = source["eightBit"];
	    }
	}
	export class ConfigData {
	    server: string;
	    port: number;
	    username: string;
	    password: string;
	    authType: string;
	    startTLS: boolean;
	    skipVerify: boolean;
	    templates: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new ConfigData(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.authType = source["authType"];
	        this.startTLS = source["startTLS"];
	        this.skipVerify = source["skipVerify"];
	        this.templates = source["templates"];
	    }
	}
	export class ServerInfo {
	    server: string;
	    port: number;
	    tlsActive: boolean;
	    authTypes: string[];

	    static createFrom(source: any = {}) {
	        return new ServerInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.port = source["port"];
	        this.tlsActive = source["tlsActive"];
	        this.authTypes = source["authTypes"];
	    }
	}
	export class ConnectionResult {
	    success: boolean;
	    message: string;
	    error?: string;
	    timestamp: string;
	    serverInfo?: ServerInfo;
	    capabilities?: CapabilityInfo;

	    static createFrom(source: any = {}) {
	        return new ConnectionResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.timestamp = source["timestamp"];
	        this.serverInfo = this.convertValues(source["serverInfo"], ServerInfo);
	        this.capabilities = this.convertValues(source["capabilities"], CapabilityInfo);
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
	export class EmailRequest {
	    config?: ConfigData;
	    from: string;
	    to: string[];
	    cc: string[];
	    bcc: string[];
	    subject: string;
	    body: string;
	    htmlBody: string;
	    attachments: string[];
	    headers: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new EmailRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], ConfigData);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.cc = source["cc"];
	        this.bcc = source["bcc"];
	        this.subject = source["subject"];
	        this.body = source["body"];
	        this.htmlBody = source["htmlBody"];
	        this.attachments = source["attachments"];
	        this.headers = source["headers"];
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
	export class SendResult {
	    success: boolean;
	    message: string;
	    error?: string;
	    timestamp: string;
	    messageId?: string;

	    static createFrom(source: any = {}) {
	        return new SendResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.timestamp = source["timestamp"];
	        this.messageId = source["messageId"];
	    }
	}

	export class TemplateData {
	    from: string;
	    to: string[];
	    cc: string[];
	    bcc: string[];
	    subject: string;
	    data: Record<string, any>;

	    static createFrom(source: any = {}) {
	        return new TemplateData(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.cc = source["cc"];
	        this.bcc = source["bcc"];
	        this.subject = source["subject"];
	        this.data = source["data"];
	    }
	}
	export class TemplateInfo {
	    name: string;
	    path: string;
	    size: number;
	    modTime: string;
	    variables: string[];
	    hasHtml: boolean;
	    hasText: boolean;
	    description: string;

	    static createFrom(source: any = {}) {
	        return new TemplateInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.modTime = source["modTime"];
	        this.variables = source["variables"];
	        this.hasHtml = source["hasHtml"];
	        this.hasText = source["hasText"];
	        this.description = source["description"];
	    }
	}
	export class TemplateResult {
	    success: boolean;
	    message: string;
	    error?: string;
	    content?: string;
	    variables?: string[];

	    static createFrom(source: any = {}) {
	        return new TemplateResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.content = source["content"];
	        this.variables = source["variables"];
	    }
	}

}
