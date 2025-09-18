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
	    current_time: StatTime;
	
	    static createFrom(source: any = {}) {
	        return new SplitPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.split_index = source["split_index"];
	        this.split_segment_id = source["split_segment_id"];
	        this.current_time = this.convertValues(source["current_time"], StatTime);
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
	export class StatTime {
	    raw: number;
	    formatted: string;
	
	    static createFrom(source: any = {}) {
	        return new StatTime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.raw = source["raw"];
	        this.formatted = source["formatted"];
	    }
	}
	export class RunPayload {
	    id: string;
	    splitfile_version: number;
	    total_time: StatTime;
	    completed: boolean;
	    split_payloads: SplitPayload[];
	
	    static createFrom(source: any = {}) {
	        return new RunPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.splitfile_version = source["splitfile_version"];
	        this.total_time = this.convertValues(source["total_time"], StatTime);
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
	export class SegmentPayload {
	    id: string;
	    name: string;
	    gold: StatTime;
	    average: StatTime;
	    pb: StatTime;
	
	    static createFrom(source: any = {}) {
	        return new SegmentPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.gold = this.convertValues(source["gold"], StatTime);
	        this.average = this.convertValues(source["average"], StatTime);
	        this.pb = this.convertValues(source["pb"], StatTime);
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
	export class WindowParams {
	    width: number;
	    height: number;
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class SplitFilePayload {
	    id: string;
	    version: number;
	    game_name: string;
	    game_category: string;
	    segments: SegmentPayload[];
	    attempts: number;
	    runs: RunPayload[];
	    SOB: StatTime;
	    window_params: WindowParams;
	
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
	        this.SOB = this.convertValues(source["SOB"], StatTime);
	        this.window_params = this.convertValues(source["window_params"], WindowParams);
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
	    current_time: StatTime;
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
	        this.current_time = this.convertValues(source["current_time"], StatTime);
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

