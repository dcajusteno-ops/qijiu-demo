export namespace main {
	
	export class AutoRuleAction {
	    type: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new AutoRuleAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.value = source["value"];
	    }
	}
	export class AutoRuleCondition {
	    field: string;
	    operator: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new AutoRuleCondition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.field = source["field"];
	        this.operator = source["operator"];
	        this.value = source["value"];
	    }
	}
	export class AutoRule {
	    id: string;
	    name: string;
	    enabled: boolean;
	    matchMode: string;
	    conditions: AutoRuleCondition[];
	    actions: AutoRuleAction[];
	    lastRunAt?: string;
	    lastMatchCount?: number;
	    lastStatus?: string;
	    lastError?: string;
	    createdAt?: string;
	    updatedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new AutoRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.matchMode = source["matchMode"];
	        this.conditions = this.convertValues(source["conditions"], AutoRuleCondition);
	        this.actions = this.convertValues(source["actions"], AutoRuleAction);
	        this.lastRunAt = source["lastRunAt"];
	        this.lastMatchCount = source["lastMatchCount"];
	        this.lastStatus = source["lastStatus"];
	        this.lastError = source["lastError"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	
	
	export class AutoRulesRunSummary {
	    totalCount: number;
	    processedCount: number;
	    matchedCount: number;
	    updatedCount: number;
	    errorCount: number;
	    ranAt: string;
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new AutoRulesRunSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCount = source["totalCount"];
	        this.processedCount = source["processedCount"];
	        this.matchedCount = source["matchedCount"];
	        this.updatedCount = source["updatedCount"];
	        this.errorCount = source["errorCount"];
	        this.ranAt = source["ranAt"];
	        this.errors = source["errors"];
	    }
	}
	export class AutoRulesStore {
	    enabled: boolean;
	    rules: AutoRule[];
	
	    static createFrom(source: any = {}) {
	        return new AutoRulesStore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.rules = this.convertValues(source["rules"], AutoRule);
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
	    order?: number;
	    enabled: boolean;
	    locked?: boolean;
	    isBuiltin?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CustomRoot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.icon = source["icon"];
	        this.order = source["order"];
	        this.enabled = source["enabled"];
	        this.locked = source["locked"];
	        this.isBuiltin = source["isBuiltin"];
	    }
	}
	export class DirectoryBinding {
	    rootDir: string;
	    outputDir: string;
	    outputRelPath: string;
	    configured: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DirectoryBinding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootDir = source["rootDir"];
	        this.outputDir = source["outputDir"];
	        this.outputRelPath = source["outputRelPath"];
	        this.configured = source["configured"];
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
	    loras?: string[];
	    searchText?: string;
	
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
	        this.loras = source["loras"];
	        this.searchText = source["searchText"];
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
	export class PromptAssistantState {
	    favoriteIds: string[];
	    recentIds: string[];
	    activeSource?: string;
	    activeCategory?: string;
	    activeSubcategory?: string;
	    activeScope?: string;
	    viewMode?: string;
	    activeEditor?: string;
	    itemsPerPage?: number;
	    currentPage?: number;
	
	    static createFrom(source: any = {}) {
	        return new PromptAssistantState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.favoriteIds = source["favoriteIds"];
	        this.recentIds = source["recentIds"];
	        this.activeSource = source["activeSource"];
	        this.activeCategory = source["activeCategory"];
	        this.activeSubcategory = source["activeSubcategory"];
	        this.activeScope = source["activeScope"];
	        this.viewMode = source["viewMode"];
	        this.activeEditor = source["activeEditor"];
	        this.itemsPerPage = source["itemsPerPage"];
	        this.currentPage = source["currentPage"];
	    }
	}
	export class PromptLibraryEntry {
	    id: string;
	    source: string;
	    category: string;
	    subcategory: string;
	    scope: string;
	    text_en: string;
	    text_zh: string;
	    preview: string;
	    extra_id: string;
	    search_text: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptLibraryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source = source["source"];
	        this.category = source["category"];
	        this.subcategory = source["subcategory"];
	        this.scope = source["scope"];
	        this.text_en = source["text_en"];
	        this.text_zh = source["text_zh"];
	        this.preview = source["preview"];
	        this.extra_id = source["extra_id"];
	        this.search_text = source["search_text"];
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
	export class UtilityMenuItem {
	    id: string;
	    visible: boolean;
	    order?: number;
	
	    static createFrom(source: any = {}) {
	        return new UtilityMenuItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.visible = source["visible"];
	        this.order = source["order"];
	    }
	}
	export class UtilityMenuState {
	    items?: UtilityMenuItem[];
	
	    static createFrom(source: any = {}) {
	        return new UtilityMenuState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], UtilityMenuItem);
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
	export class UserProfile {
	    displayName?: string;
	    headline?: string;
	    bio?: string;
	    location?: string;
	    website?: string;
	    dailyGoal?: number;
	    preferredStartPage?: string;
	    imagePath?: string;
	
	    static createFrom(source: any = {}) {
	        return new UserProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.headline = source["headline"];
	        this.bio = source["bio"];
	        this.location = source["location"];
	        this.website = source["website"];
	        this.dailyGoal = source["dailyGoal"];
	        this.preferredStartPage = source["preferredStartPage"];
	        this.imagePath = source["imagePath"];
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
	    outputConfigured?: boolean;
	    pathVersion?: number;
	    shortcutSettings?: ShortcutSettings;
	    userProfile?: UserProfile;
	    utilityMenu?: UtilityMenuState;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trashRetentionDays = source["trashRetentionDays"];
	        this.rootDir = source["rootDir"];
	        this.outputDir = source["outputDir"];
	        this.outputConfigured = source["outputConfigured"];
	        this.pathVersion = source["pathVersion"];
	        this.shortcutSettings = this.convertValues(source["shortcutSettings"], ShortcutSettings);
	        this.userProfile = this.convertValues(source["userProfile"], UserProfile);
	        this.utilityMenu = this.convertValues(source["utilityMenu"], UtilityMenuState);
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

