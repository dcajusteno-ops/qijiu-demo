export namespace main {
	
	export class CacheClearResult {
	    deletedFiles: number;
	    deletedDirs: number;
	    bytesFreed: number;
	
	    static createFrom(source: any = {}) {
	        return new CacheClearResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deletedFiles = source["deletedFiles"];
	        this.deletedDirs = source["deletedDirs"];
	        this.bytesFreed = source["bytesFreed"];
	    }
	}
	export class CustomRoot {
	    id: string;
	    name: string;
	    path: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomRoot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.icon = source["icon"];
	    }
	}
	export class DirectoryBinding {
	    rootDir: string;
	    outputDir: string;
	    outputRelPath: string;
	
	    static createFrom(source: any = {}) {
	        return new DirectoryBinding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootDir = source["rootDir"];
	        this.outputDir = source["outputDir"];
	        this.outputRelPath = source["outputRelPath"];
	    }
	}
	export class FavoriteGroup {
	    id: string;
	    name: string;
	    paths: string[];
	
	    static createFrom(source: any = {}) {
	        return new FavoriteGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.paths = source["paths"];
	    }
	}
	export class ImageFile {
	    name: string;
	    path: string;
	    relPath: string;
	    modTime: string;
	    size: number;
	    width: number;
	    height: number;
	    prompt?: string;
	    model?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.relPath = source["relPath"];
	        this.modTime = source["modTime"];
	        this.size = source["size"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.prompt = source["prompt"];
	        this.model = source["model"];
	    }
	}
	export class ImageMetadata {
	    relPath: string;
	    format: string;
	    width: number;
	    height: number;
	    hasMetadata: boolean;
	    prompt: string;
	    workflow: string;
	    positive: string;
	    negative: string;
	    model: string;
	    sampler: string;
	    scheduler: string;
	    seed: string;
	    steps: string;
	    cfg: string;
	    loras: string[];
	    nodeCount: number;
	    extraFields: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ImageMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relPath = source["relPath"];
	        this.format = source["format"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.hasMetadata = source["hasMetadata"];
	        this.prompt = source["prompt"];
	        this.workflow = source["workflow"];
	        this.positive = source["positive"];
	        this.negative = source["negative"];
	        this.model = source["model"];
	        this.sampler = source["sampler"];
	        this.scheduler = source["scheduler"];
	        this.seed = source["seed"];
	        this.steps = source["steps"];
	        this.cfg = source["cfg"];
	        this.loras = source["loras"];
	        this.nodeCount = source["nodeCount"];
	        this.extraFields = source["extraFields"];
	    }
	}
	export class LauncherTool {
	    id: string;
	    name: string;
	    path: string;
	    icon: string;
	    args: string;
	
	    static createFrom(source: any = {}) {
	        return new LauncherTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.icon = source["icon"];
	        this.args = source["args"];
	    }
	}
	export class PromptTemplate {
	    id: string;
	    name: string;
	    content: string;
	    type: string;
	    category: string;
	    sourcePath: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.content = source["content"];
	        this.type = source["type"];
	        this.category = source["category"];
	        this.sourcePath = source["sourcePath"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class PromptToolLink {
	    id: string;
	    name: string;
	    url: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptToolLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.icon = source["icon"];
	    }
	}
	export class ShortcutBinding {
	    action: string;
	    accelerator: string;
	
	    static createFrom(source: any = {}) {
	        return new ShortcutBinding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.accelerator = source["accelerator"];
	    }
	}
	export class ShortcutSettings {
	    enabled: boolean;
	    bindings: ShortcutBinding[];
	
	    static createFrom(source: any = {}) {
	        return new ShortcutSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.bindings = this.convertValues(source["bindings"], ShortcutBinding);
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
	export class Settings {
	    trashRetentionDays: number;
	    rootDir?: string;
	    outputDir?: string;
	    pathVersion?: number;
	    shortcutSettings?: ShortcutSettings;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trashRetentionDays = source["trashRetentionDays"];
	        this.rootDir = source["rootDir"];
	        this.outputDir = source["outputDir"];
	        this.pathVersion = source["pathVersion"];
	        this.shortcutSettings = this.convertValues(source["shortcutSettings"], ShortcutSettings);
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
	export class ShortcutAction {
	    id: string;
	    label: string;
	    description: string;
	    defaultAccelerator: string;
	
	    static createFrom(source: any = {}) {
	        return new ShortcutAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.defaultAccelerator = source["defaultAccelerator"];
	    }
	}
	
	
	export class SmartAlbum {
	    field: string;
	    value: string;
	    count: number;
	    paths: string[];
	
	    static createFrom(source: any = {}) {
	        return new SmartAlbum(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.field = source["field"];
	        this.value = source["value"];
	        this.count = source["count"];
	        this.paths = source["paths"];
	    }
	}
	export class SmartAlbumField {
	    key: string;
	    label: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new SmartAlbumField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	    }
	}
	export class Stats {
	    totalCount: number;
	    todayCount: number;
	    totalSize: number;
	    byDate: Record<string, number>;
	    byTag: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCount = source["totalCount"];
	        this.todayCount = source["todayCount"];
	        this.totalSize = source["totalSize"];
	        this.byDate = source["byDate"];
	        this.byTag = source["byTag"];
	    }
	}
	export class Tag {
	    id: string;
	    name: string;
	    color: string;
	    category: string;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	        this.category = source["category"];
	    }
	}
	export class TrashItem {
	    filename: string;
	    originalPath: string;
	    deletedAt: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new TrashItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.originalPath = source["originalPath"];
	        this.deletedAt = source["deletedAt"];
	        this.path = source["path"];
	    }
	}
	export class UploadResult {
	    count: number;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new UploadResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.errors = source["errors"];
	    }
	}

}

