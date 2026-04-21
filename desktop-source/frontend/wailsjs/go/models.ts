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
	    pinned?: boolean;
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
	        this.pinned = source["pinned"];
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
	export class DirectoryHealthIssue {
	    key: string;
	    level: string;
	    title: string;
	    description: string;
	    count: number;
	    action?: string;
	
	    static createFrom(source: any = {}) {
	        return new DirectoryHealthIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.level = source["level"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.count = source["count"];
	        this.action = source["action"];
	    }
	}
	export class DirectoryHealthSummary {
	    totalImages: number;
	    emptyFolderCount: number;
	    invalidTagReferenceCount: number;
	    invalidFavoriteReferenceCount: number;
	    thumbCacheCount: number;
	    thumbCacheBytes: number;
	    previewCacheCount: number;
	    previewCacheBytes: number;
	    issues: DirectoryHealthIssue[];
	
	    static createFrom(source: any = {}) {
	        return new DirectoryHealthSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalImages = source["totalImages"];
	        this.emptyFolderCount = source["emptyFolderCount"];
	        this.invalidTagReferenceCount = source["invalidTagReferenceCount"];
	        this.invalidFavoriteReferenceCount = source["invalidFavoriteReferenceCount"];
	        this.thumbCacheCount = source["thumbCacheCount"];
	        this.thumbCacheBytes = source["thumbCacheBytes"];
	        this.previewCacheCount = source["previewCacheCount"];
	        this.previewCacheBytes = source["previewCacheBytes"];
	        this.issues = this.convertValues(source["issues"], DirectoryHealthIssue);
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
	export class GalleryPerformanceSettings {
	    mode: string;
	    initialBatchSize: number;
	    pageSize: number;
	    thumbPreferred: boolean;
	    backgroundVariantWarmup: boolean;
	    metadataLazy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GalleryPerformanceSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.initialBatchSize = source["initialBatchSize"];
	        this.pageSize = source["pageSize"];
	        this.thumbPreferred = source["thumbPreferred"];
	        this.backgroundVariantWarmup = source["backgroundVariantWarmup"];
	        this.metadataLazy = source["metadataLazy"];
	    }
	}
	export class GetImagesPageQuery {
	    sortBy: string;
	    sortOrder: string;
	    page: number;
	    pageSize: number;
	    scopeRelPath?: string;
	    favoritesOnly?: boolean;
	    favoriteGroupId?: string;
	    searchQuery?: string;
	    activeTagId?: string;
	    activeModelFilter?: string;
	    activeLoraFilter?: string;
	    activeDatePreset?: string;
	    activeDateStart?: string;
	    activeDateEnd?: string;
	
	    static createFrom(source: any = {}) {
	        return new GetImagesPageQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sortBy = source["sortBy"];
	        this.sortOrder = source["sortOrder"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.scopeRelPath = source["scopeRelPath"];
	        this.favoritesOnly = source["favoritesOnly"];
	        this.favoriteGroupId = source["favoriteGroupId"];
	        this.searchQuery = source["searchQuery"];
	        this.activeTagId = source["activeTagId"];
	        this.activeModelFilter = source["activeModelFilter"];
	        this.activeLoraFilter = source["activeLoraFilter"];
	        this.activeDatePreset = source["activeDatePreset"];
	        this.activeDateStart = source["activeDateStart"];
	        this.activeDateEnd = source["activeDateEnd"];
	    }
	}
	export class ImageFile {
	    name: string;
	    path: string;
	    thumbPath?: string;
	    previewPath?: string;
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
	        this.thumbPath = source["thumbPath"];
	        this.previewPath = source["previewPath"];
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
	export class GetImagesPageResult {
	    items: ImageFile[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	    hasMore: boolean;
	    mode: string;
	    modeReason?: string;
	
	    static createFrom(source: any = {}) {
	        return new GetImagesPageResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], ImageFile);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
	        this.hasMore = source["hasMore"];
	        this.mode = source["mode"];
	        this.modeReason = source["modeReason"];
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
	
	export class ImageGallerySummary {
	    totalImages: number;
	    managedRootCount: number;
	    activeMode: string;
	    modeReason: string;
	    thumbCacheCount: number;
	    thumbCacheBytes: number;
	    previewCacheCount: number;
	    previewCacheBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new ImageGallerySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalImages = source["totalImages"];
	        this.managedRootCount = source["managedRootCount"];
	        this.activeMode = source["activeMode"];
	        this.modeReason = source["modeReason"];
	        this.thumbCacheCount = source["thumbCacheCount"];
	        this.thumbCacheBytes = source["thumbCacheBytes"];
	        this.previewCacheCount = source["previewCacheCount"];
	        this.previewCacheBytes = source["previewCacheBytes"];
	    }
	}
	export class PromptCandidateDebug {
	    text: string;
	    score: number;
	    sourceNodeId?: string;
	    sourceClass?: string;
	    sourceTitle?: string;
	    sourceKey?: string;
	    strategy?: string;
	    depth?: number;
	
	    static createFrom(source: any = {}) {
	        return new PromptCandidateDebug(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.score = source["score"];
	        this.sourceNodeId = source["sourceNodeId"];
	        this.sourceClass = source["sourceClass"];
	        this.sourceTitle = source["sourceTitle"];
	        this.sourceKey = source["sourceKey"];
	        this.strategy = source["strategy"];
	        this.depth = source["depth"];
	    }
	}
	export class PromptSelectionDebug {
	    selectedText?: string;
	    strategy?: string;
	    sourceNodeId?: string;
	    sourceClass?: string;
	    sourceTitle?: string;
	    sourceKey?: string;
	    candidates?: PromptCandidateDebug[];
	
	    static createFrom(source: any = {}) {
	        return new PromptSelectionDebug(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selectedText = source["selectedText"];
	        this.strategy = source["strategy"];
	        this.sourceNodeId = source["sourceNodeId"];
	        this.sourceClass = source["sourceClass"];
	        this.sourceTitle = source["sourceTitle"];
	        this.sourceKey = source["sourceKey"];
	        this.candidates = this.convertValues(source["candidates"], PromptCandidateDebug);
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
	export class PromptDebugInfo {
	    positive: PromptSelectionDebug;
	    negative: PromptSelectionDebug;
	
	    static createFrom(source: any = {}) {
	        return new PromptDebugInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.positive = this.convertValues(source["positive"], PromptSelectionDebug);
	        this.negative = this.convertValues(source["negative"], PromptSelectionDebug);
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
	    promptDebug?: PromptDebugInfo;
	
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
	        this.promptDebug = this.convertValues(source["promptDebug"], PromptDebugInfo);
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
	    galleryPerformanceMode?: string;
	    galleryInitialBatchSize?: number;
	    galleryPageSize?: number;
	    galleryThumbPreferred?: boolean;
	    galleryBackgroundVariantWarmup?: boolean;
	    galleryMetadataLazy?: boolean;
	
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
	        this.galleryPerformanceMode = source["galleryPerformanceMode"];
	        this.galleryInitialBatchSize = source["galleryInitialBatchSize"];
	        this.galleryPageSize = source["galleryPageSize"];
	        this.galleryThumbPreferred = source["galleryThumbPreferred"];
	        this.galleryBackgroundVariantWarmup = source["galleryBackgroundVariantWarmup"];
	        this.galleryMetadataLazy = source["galleryMetadataLazy"];
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
	
	
	
	export class WorkbenchRecentDate {
	    date: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkbenchRecentDate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.count = source["count"];
	    }
	}
	export class WorkbenchSummary {
	    total: number;
	    datedTotal: number;
	    today: number;
	    yesterday: number;
	    last7: number;
	    month: number;
	    recentDates: WorkbenchRecentDate[];
	
	    static createFrom(source: any = {}) {
	        return new WorkbenchSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.datedTotal = source["datedTotal"];
	        this.today = source["today"];
	        this.yesterday = source["yesterday"];
	        this.last7 = source["last7"];
	        this.month = source["month"];
	        this.recentDates = this.convertValues(source["recentDates"], WorkbenchRecentDate);
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
	export class WorkbenchFilterOption {
	    value: string;
	    label: string;
	    count: number;
	    aliases?: string[];
	
	    static createFrom(source: any = {}) {
	        return new WorkbenchFilterOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.label = source["label"];
	        this.count = source["count"];
	        this.aliases = source["aliases"];
	    }
	}
	export class WorkbenchAggregateResult {
	    availableModels: WorkbenchFilterOption[];
	    availableLoras: WorkbenchFilterOption[];
	    summary: WorkbenchSummary;
	    filteredCount: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkbenchAggregateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.availableModels = this.convertValues(source["availableModels"], WorkbenchFilterOption);
	        this.availableLoras = this.convertValues(source["availableLoras"], WorkbenchFilterOption);
	        this.summary = this.convertValues(source["summary"], WorkbenchSummary);
	        this.filteredCount = source["filteredCount"];
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
	
	
	
	export class WorkbenchSummaryQuery {
	    activeDatePreset?: string;
	    activeDateStart?: string;
	    activeDateEnd?: string;
	    activeModelFilter?: string;
	    activeLoraFilter?: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkbenchSummaryQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activeDatePreset = source["activeDatePreset"];
	        this.activeDateStart = source["activeDateStart"];
	        this.activeDateEnd = source["activeDateEnd"];
	        this.activeModelFilter = source["activeModelFilter"];
	        this.activeLoraFilter = source["activeLoraFilter"];
	    }
	}

}

