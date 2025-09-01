export namespace session {
	
	export class SplitFile {
	
	
	    static createFrom(source: any = {}) {
	        return new SplitFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

