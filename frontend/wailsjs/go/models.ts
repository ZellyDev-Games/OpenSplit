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
	export class SplitPayload {
	    split_index: number;
	    split_segment_id: string;
	    current_time: string;
	    current_duration: number;
	
	    static createFrom(source: any = {}) {
	        return new SplitPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.split_index = source["split_index"];
	        this.split_segment_id = source["split_segment_id"];
	        this.current_time = source["current_time"];
	        this.current_duration = source["current_duration"];
	    }
	}
	export class RunPayload {
	    id: number[];
	    splitfile_version: number;
	    total_time: number;
	    completed: boolean;
	    split_payloads: SplitPayload[];
	
	    static createFrom(source: any = {}) {
	        return new RunPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.splitfile_version = source["splitfile_version"];
	        this.total_time = source["total_time"];
	        this.completed = source["completed"];
	        this.split_payloads = this.convertValues(source["split_payloads"], SplitPayload);
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
	export class PBStatsPayload {
	    run?: RunPayload;
	    total: string;
	
	    static createFrom(source: any = {}) {
	        return new PBStatsPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.run = this.convertValues(source["run"], RunPayload);
	        this.total = source["total"];
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
	export class StatTimePayload {
	    id: string;
	    time: string;
	
	    static createFrom(source: any = {}) {
	        return new StatTimePayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.time = source["time"];
	    }
	}
	export class SplitFileStatsPayload {
	    golds: StatTimePayload[];
	    averages: StatTimePayload[];
	    sob: string;
	    pb?: PBStatsPayload;
	
	    static createFrom(source: any = {}) {
	        return new SplitFileStatsPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.golds = this.convertValues(source["golds"], StatTimePayload);
	        this.averages = this.convertValues(source["averages"], StatTimePayload);
	        this.sob = source["sob"];
	        this.pb = this.convertValues(source["pb"], PBStatsPayload);
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
	export class SplitFilePayload {
	    id: number[];
	    version: number;
	    game_name: string;
	    game_category: string;
	    segments: SegmentPayload[];
	    attempts: number;
	    runs: RunPayload[];
	    stats: SplitFileStatsPayload;
	
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
	        this.stats = this.convertValues(source["stats"], SplitFileStatsPayload);
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
	    current_run?: RunPayload;
	
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
	        this.current_run = this.convertValues(source["current_run"], RunPayload);
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

