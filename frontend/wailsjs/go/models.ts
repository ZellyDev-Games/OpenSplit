export namespace statemachine {
	
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

