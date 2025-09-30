export namespace dispatcher {
	
	export class DispatchReply {
	    code: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DispatchReply(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}

}

export namespace hotkeys {
	
	export class KeyInfo {
	    key_code: number;
	    locale_name: string;
	
	    static createFrom(source: any = {}) {
	        return new KeyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key_code = source["key_code"];
	        this.locale_name = source["locale_name"];
	    }
	}

}

