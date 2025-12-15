export namespace app {
	
	export class QueryResult {
	    success: boolean;
	    data: any;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.data = source["data"];
	        this.error = source["error"];
	    }
	}

}

export namespace dto {
	
	export class ChatMessage {
	    role: string;
	    content: string;
	    timestamp?: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class ConversationInfo {
	    id: string;
	    title: string;
	    preview: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversationInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.preview = source["preview"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ExcelStatus {
	    connected: boolean;
	    workbooks: excel.Workbook[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExcelStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.workbooks = this.convertValues(source["workbooks"], excel.Workbook);
	        this.error = source["error"];
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
	export class ModelInfo {
	    id: string;
	    name: string;
	    description: string;
	    contextLength: number;
	    pricePrompt: string;
	    priceComplete: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.contextLength = source["contextLength"];
	        this.pricePrompt = source["pricePrompt"];
	        this.priceComplete = source["priceComplete"];
	    }
	}
	export class PreviewData {
	    headers: string[];
	    rows: string[][];
	    totalRows: number;
	    totalCols: number;
	    workbook: string;
	    sheet: string;
	
	    static createFrom(source: any = {}) {
	        return new PreviewData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.headers = source["headers"];
	        this.rows = source["rows"];
	        this.totalRows = source["totalRows"];
	        this.totalCols = source["totalCols"];
	        this.workbook = source["workbook"];
	        this.sheet = source["sheet"];
	    }
	}

}

export namespace excel {
	
	export class CellData {
	    row: number;
	    col: number;
	    value: any;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new CellData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.row = source["row"];
	        this.col = source["col"];
	        this.value = source["value"];
	        this.text = source["text"];
	    }
	}
	export class SheetData {
	    name: string;
	    headers: string[];
	    rows: CellData[][];
	
	    static createFrom(source: any = {}) {
	        return new SheetData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.headers = source["headers"];
	        this.rows = this.convertValues(source["rows"], CellData);
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
	export class Workbook {
	    name: string;
	    path: string;
	    sheets: string[];
	
	    static createFrom(source: any = {}) {
	        return new Workbook(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.sheets = source["sheets"];
	    }
	}

}

export namespace license {
	
	export class LicenseStatus {
	    valid: boolean;
	    message: string;
	    hash?: string;
	    // Go type: time
	    activatedAt?: any;
	    machineId?: string;
	
	    static createFrom(source: any = {}) {
	        return new LicenseStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.message = source["message"];
	        this.hash = source["hash"];
	        this.activatedAt = this.convertValues(source["activatedAt"], null);
	        this.machineId = source["machineId"];
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

export namespace storage {
	
	export class ProviderConfig {
	    apiKey?: string;
	    model?: string;
	    baseUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.baseUrl = source["baseUrl"];
	    }
	}
	export class Config {
	    provider?: string;
	    apiKey?: string;
	    model: string;
	    baseUrl?: string;
	    providerConfigs?: Record<string, ProviderConfig>;
	    maxRowsContext: number;
	    maxRowsPreview: number;
	    includeHeaders: boolean;
	    detailLevel: string;
	    customPrompt: string;
	    language: string;
	    lastUsedWorkbook?: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.baseUrl = source["baseUrl"];
	        this.providerConfigs = this.convertValues(source["providerConfigs"], ProviderConfig, true);
	        this.maxRowsContext = source["maxRowsContext"];
	        this.maxRowsPreview = source["maxRowsPreview"];
	        this.includeHeaders = source["includeHeaders"];
	        this.detailLevel = source["detailLevel"];
	        this.customPrompt = source["customPrompt"];
	        this.language = source["language"];
	        this.lastUsedWorkbook = source["lastUsedWorkbook"];
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

