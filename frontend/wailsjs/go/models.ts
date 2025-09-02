export namespace session {
	
	export class SegmentPayload {
	    id: number[];
	    name: string;
	    best_time: number;
	    average_time: number;
	
	    static createFrom(source: any = {}) {
	        return new SegmentPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.best_time = source["best_time"];
	        this.average_time = source["average_time"];
	    }
	}
	export class SplitFilePayload {
	    game_name: string;
	    game_category: string;
	    segments: SegmentPayload[];
	    attempts: number;
	
	    static createFrom(source: any = {}) {
	        return new SplitFilePayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_name = source["game_name"];
	        this.game_category = source["game_category"];
	        this.segments = this.convertValues(source["segments"], SegmentPayload);
	        this.attempts = source["attempts"];
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

