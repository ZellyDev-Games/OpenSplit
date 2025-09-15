export namespace session {
	
	export class Config {
	    speed_run_API_base: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.speed_run_API_base = source["speed_run_API_base"];
	    }
	}
	export class RunPayload {
	    id: number[];
	    splitFileId: number[];
	    splitFileVersion: number;
	
	    static createFrom(source: any = {}) {
	        return new RunPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.splitFileId = source["splitFileId"];
	        this.splitFileVersion = source["splitFileVersion"];
	    }
	}
	export class SegmentPayload {
	    id: string;
	    name: string;
	    best_time: string;
	    average_time: string;
	
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
	    id: number[];
	    version: number;
	    game_name: string;
	    game_category: string;
	    segments: SegmentPayload[];
	    attempts: number;
	    runs: RunPayload[];
	
	    static createFrom(source: any = {}) {
	        return new SplitFilePayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version = source["version"];
	        this.game_name = source["game_name"];
	        this.game_category = source["game_category"];
	        this.segments = this.convertValues(source["segments"], SegmentPayload);
	        this.attempts = source["attempts"];
	        this.runs = this.convertValues(source["runs"], RunPayload);
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
	export class ServicePayload {
	    split_file?: SplitFilePayload;
	    current_segment_index: number;
	    current_segment?: SegmentPayload;
	    finished: boolean;
	    paused: boolean;
	    current_time: number;
	    current_time_formatted: string;
	
	    static createFrom(source: any = {}) {
	        return new ServicePayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.split_file = this.convertValues(source["split_file"], SplitFilePayload);
	        this.current_segment_index = source["current_segment_index"];
	        this.current_segment = this.convertValues(source["current_segment"], SegmentPayload);
	        this.finished = source["finished"];
	        this.paused = source["paused"];
	        this.current_time = source["current_time"];
	        this.current_time_formatted = source["current_time_formatted"];
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

