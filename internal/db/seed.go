package db

// seedData 向数据库写入初始数据（分类、作物/场所、防治对象）。
// 所有 INSERT 均使用 INSERT OR IGNORE，保证重复启动不会报错也不会重复插入。
// 产品数据不在此处预置，由管理员通过管理台录入。
func seedData() error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// ──────────────────────────────────────────────────────────────────────────
	// 1. 分类数据（categories）
	//    规则：parent_id=0 为一级，parent_id>0 指向一级 id
	//    注意：SQLite AUTOINCREMENT 从 1 开始，下面用具体 id 插入以保持一致性
	// ──────────────────────────────────────────────────────────────────────────
	categories := []struct {
		id       int
		parentID int
		name     string
	}{
		// 一级分类
		{1, 0, "除草"},
		{2, 0, "虫害"},
		{3, 0, "病害"},
		{4, 0, "调节剂"},
		{5, 0, "拌种剂"},
		// 除草 二级分类
		{6, 1, "灭生性除草剂"},
		{7, 1, "封闭除草剂"},
		{8, 1, "玉米除草剂"},
		{9, 1, "小麦除草剂"},
		{10, 1, "夏季除草剂单剂"},
		{11, 1, "夏季除草剂复配系列"},
		// 杀虫 二级分类
		{12, 2, "地下害虫用药"},
		{13, 2, "青虫/肉虫系列"},
		// 杀菌 二级分类（暂无，使用一级 id=3）
		// 植物生长调节剂 二级（暂无）
		{14, 5, "花生拌种剂"},
		{15, 5, "小麦拌种剂"},
		{16, 4, "叶面肥系列"},
	}

	catSQL := `INSERT OR IGNORE INTO categories (id, parent_id, name, sort_order) VALUES (?, ?, ?, ?)`
	for i, c := range categories {
		if _, err = tx.Exec(catSQL, c.id, c.parentID, c.name, i); err != nil {
			return err
		}
	}

	// ──────────────────────────────────────────────────────────────────────────
	// 2. 作物/场所数据（crops）
	//    统一存储：农作物 + 除草类"使用场所"均在此表
	// ──────────────────────────────────────────────────────────────────────────
	crops := []string{
		// 农作物（主粮/经济作物）
		"大豆", "小麦", "玉米", "水稻", "棉花", "花生", "甘蔗", "油菜",
		"高粱", "谷子", "向日葵", "甜菜", "芝麻", "大麦", "烟草", "荞麦",
		// 果菜类
		"辣椒", "番茄", "茄子", "黄瓜", "南瓜", "冬瓜", "丝瓜", "苦瓜",
		"西葫芦", "青椒", "樱桃番茄", "西红柿",
		// 叶菜类
		"白菜", "甘蓝", "花椰菜", "菠菜", "生菜", "莴苣", "芹菜", "茴香", "香菜",
		// 根茎类
		"土豆", "萝卜", "胡萝卜", "洋葱", "大蒜", "生姜", "山药", "魔芋",
		// 豆类
		"四季豆", "豇豆", "豌豆", "蚕豆", "韭黄",
		// 葱蒜类
		"韭菜", "葱", "姜", "芥菜",
		// 果树/经济作物
		"苹果", "梨", "桃", "杏", "李", "樱桃", "葡萄", "草莓",
		"柑橘", "柠檬", "荔枝", "龙眼", "芒果", "香蕉", "菠萝",
		"西瓜", "甜瓜", "哈密瓜",
		// 水稻专用场景（新增分类）
		"移栽水稻", "水稻秧田", "水稻本田",
		// 除草场所
		"果园", "茶园", "橡胶园", "田埂", "铁路沿线", "荒山荒地",
		"绿化园林", "工业园区", "草坪", "路边",
	}

	cropSQL := `INSERT OR IGNORE INTO crops (name, sort_order) VALUES (?, ?)`
	for i, name := range crops {
		if _, err = tx.Exec(cropSQL, name, i); err != nil {
			return err
		}
	}

	// ──────────────────────────────────────────────────────────────────────────
	// 3. 防治对象数据（targets）
	//    按 type 分四类：weed(杂草) / pest(害虫) / disease(病害) / regulator(调节剂)
	// ──────────────────────────────────────────────────────────────────────────
	type targetSeed struct {
		name string
		typ  string
	}
	targets := []targetSeed{
		// 杂草 (weed)
		{"藜", "weed"}, {"蓼", "weed"}, {"苍耳", "weed"}, {"茅草", "weed"}, {"马齿苋", "weed"},
		{"田旋花", "weed"}, {"打碗花", "weed"}, {"拉拉秧", "weed"}, {"猪殃殃", "weed"},
		{"灰灰菜", "weed"}, {"芦苇", "weed"}, {"稗草", "weed"}, {"马唐", "weed"}, {"狗尾草", "weed"},
		{"香附子", "weed"}, {"野燕麦", "weed"}, {"看麦娘", "weed"}, {"小飞蓬", "weed"},
		{"水花生", "weed"}, {"刺儿菜", "weed"}, {"鸭跖草", "weed"}, {"铁芒箕", "weed"},
		{"狗牙根", "weed"}, {"竹子", "weed"}, {"大树", "weed"}, {"灌木", "weed"}, {"乔木", "weed"},
		{"大部分禾本科", "weed"}, {"阔叶杂草", "weed"}, {"苋菜", "weed"}, {"龙葵", "weed"},
		{"荠菜", "weed"}, {"牛筋草", "weed"}, {"菟丝子", "weed"}, {"繁缕", "weed"},
		{"早熟禾", "weed"}, {"黑麦草", "weed"}, {"虎尾草", "weed"}, {"反枝苋", "weed"},
		{"繁缕", "weed"}, {"播娘蒿", "weed"}, {"婆婆纳", "weed"},
		{"千金子", "weed"}, {"野苋菜", "weed"}, {"车前草", "weed"}, {"铁苋菜", "weed"},
		{"眼子菜", "weed"}, {"牛毛草", "weed"}, {"异形莎草", "weed"}, {"茼麻", "weed"},
		{"碎米沙草", "weed"}, {"白茅", "weed"}, {"假高粱", "weed"}, {"鬼针草", "weed"}, {"自生麦苗", "weed"},
		{"蒲公英", "weed"}, {"苦苣菜", "weed"}, {"蓟", "weed"}, {"三叶草", "weed"}, {"野豌豆", "weed"},
		{"泽漆", "weed"}, {"麦家公", "weed"}, {"牛繁缕", "weed"}, {"宝盖草", "weed"},
		{"大巢菜", "weed"}, {"节节麦", "weed"}, {"雀麦", "weed"}, {"棒头草", "weed"},
		{"碎米荠", "weed"}, {"野油菜", "weed"}, {"野老鹳", "weed"}, {"硬草", "weed"},
		{"日本看麦娘", "weed"}, {"多花黑麦草", "weed"}, {"罔草", "weed"},

		// 害虫 (pest)
		{"蚜虫", "pest"}, {"飞虱", "pest"}, {"蓟马", "pest"}, {"粉虱", "pest"},
		{"叶蝉", "pest"}, {"稻象甲", "pest"}, {"潜叶蝇", "pest"}, {"潜叶蛾", "pest"},
		{"跳甲", "pest"}, {"稻螟虫", "pest"}, {"稻飞虱", "pest"}, {"褐飞虱", "pest"},
		{"盲蝽", "pest"}, {"木虱", "pest"}, {"茶小绿叶蝉", "pest"}, {"介壳虫", "pest"},
		{"矢尖蚧", "pest"}, {"黑刺粉虱", "pest"}, {"锈璧虱", "pest"}, {"白粉虱", "pest"},
		{"梨木虱", "pest"}, {"稻飞虱若虫", "pest"}, {"烟粉虱", "pest"}, {"蚊蝇幼虫", "pest"},
		{"桃蚜", "pest"}, {"瓜蚜", "pest"}, {"棉蚜", "pest"}, {"菜青虫", "pest"},
		{"小菜蛾", "pest"}, {"斜纹夜蛾", "pest"}, {"红蜘蛛", "pest"}, {"食心虫", "pest"},
		{"盲蝽象", "pest"}, {"茶尺蠖", "pest"}, {"蟑螂", "pest"}, {"蚂蚁", "pest"},
		{"蚊蝇", "pest"}, {"跳蚤", "pest"}, {"蛴螬", "pest"}, {"稻纵卷叶螟", "pest"},
		{"二化螟", "pest"}, {"粘虫", "pest"}, {"甜菜夜蛾", "pest"}, {"桃小食心虫", "pest"},
		{"梨小食心虫", "pest"}, {"棉铃虫", "pest"}, {"红铃虫", "pest"}, {"臭虫", "pest"},
		{"三化螟", "pest"}, {"稻苞虫", "pest"}, {"玉米螟", "pest"}, {"草地贪夜蛾", "pest"},
		{"豆荚螟", "pest"}, {"卷叶蛾", "pest"}, {"蔗螟", "pest"}, {"条螟", "pest"},
		{"茶毛虫", "pest"}, {"斑潜蝇", "pest"}, {"叶螨", "pest"}, {"松毛虫", "pest"},
		{"荔枝蒂蛀虫", "pest"}, {"苹果食心虫", "pest"}, {"果蝇", "pest"}, {"地老虎", "pest"},
		{"金针虫", "pest"}, {"象甲", "pest"}, {"金龟子", "pest"}, {"蝼蛄", "pest"},
		{"麦蜘蛛", "pest"}, {"烟青虫", "pest"}, {"绿盲蝽", "pest"}, {"蒜蛆", "pest"},
		{"根蛆", "pest"},

		// 病害 (disease)
		{"白粉病", "disease"}, {"锈病", "disease"}, {"赤霉病", "disease"},
		{"纹枯病", "disease"}, {"炭疽病", "disease"}, {"黑星病", "disease"},
		{"叶斑病", "disease"}, {"茎基腐病", "disease"}, {"白绢病", "disease"},
		{"稻曲病", "disease"}, {"褐斑病", "disease"}, {"脂点黄斑病", "disease"},
		{"斑点落叶病", "disease"}, {"跳斑病", "disease"}, {"稻瘟病", "disease"},
		{"早疫病", "disease"}, {"晚疫病", "disease"}, {"大斑病", "disease"},
		{"灰斑病", "disease"}, {"恶苗病", "disease"}, {"烂秧病", "disease"},
		{"条纹病", "disease"}, {"黄萎病", "disease"}, {"立枯病", "disease"},
		{"霜霉病", "disease"}, {"枯萎病", "disease"}, {"疫病", "disease"},
		{"青枯病", "disease"}, {"软腐病", "disease"}, {"溃疡病", "disease"},
		{"蒂腐病", "disease"}, {"黑痘病", "disease"}, {"白腐病", "disease"},
		{"根腐病", "disease"}, {"猝倒病", "disease"}, {"烟煤病", "disease"},
		{"流胶病", "disease"}, {"腐烂病", "disease"}, {"果树青苔", "disease"},
		{"叶霉病", "disease"}, {"蔓枯病", "disease"}, {"斑枯病", "disease"},
		{"轮纹病", "disease"}, {"抗性白粉病", "disease"},
		// 调节剂 (regulator)
		{"促进生长", "regulator"}, {"提高抗逆性", "regulator"}, {"促进生长", "regulator"},
		{"提高抗逆性", "regulator"}, {"增加产量", "regulator"}, {"改善品质", "regulator"},
		{"促进细胞分裂和伸长", "regulator"}, {"增强光合作用", "regulator"}, {"提高作物抗寒能力", "regulator"},
		{"提高抗旱能力", "regulator"}, {"提高抗盐碱能力", "regulator"}, {"保花保果", "regulator"},
		{"提高坐果率", "regulator"}, {"促进果实膨大", "regulator"}, {"增加糖分含量", "regulator"},
		{"使叶片浓绿厚大", "regulator"}, {"植株健壮", "regulator"}, {"缩短基部节间", "regulator"},
		{"增粗茎秆", "regulator"}, {"降低穗位高度", "regulator"}, {"增强抗倒伏能力", "regulator"},
		{"促进气生根生长", "regulator"}, {"提高光合效率", "regulator"}, {"抑制节间伸长", "regulator"},
		{"增强茎秆韧性", "regulator"}, {"降低倒伏风险", "regulator"}, {"提高叶绿素含量", "regulator"},
		{"增强养分积累", "regulator"}, {"促进果实成熟和着色", "regulator"}, {"促进根系发育", "regulator"},
		{"控制徒长", "regulator"}, {"延缓营养生长", "regulator"}, {"促进生殖生长", "regulator"},
		{"控旺", "regulator"}, {"矮化植株", "regulator"}, {"增产作用", "regulator"}, {"抑制茎秆伸长", "regulator"},
		{"增强茎秆粗度", "regulator"}, {"控制新梢旺长", "regulator"}, {"减少营养浪费", "regulator"},
		{"促进花芽分化", "regulator"}, {"增产提质", "regulator"}, {"补充磷钾", "regulator"},
		{"促进糖份积累", "regulator"}, {"改善口感", "regulator"}, {"增强叶片品质", "regulator"},
		{"促进分蘖", "regulator"}, {"促进籽粒饱满", "regulator"}, {"提高结荚率", "regulator"}, {"减少落花落果", "regulator"},
		{"改善果实甜度", "regulator"}, {"改善果实色泽", "regulator"}, {"提高抗病性", "regulator"},
		{"促进根系生长", "regulator"}, {"增强花粉活力", "regulator"}, {"提高授粉率", "regulator"},
		{"减少裂果", "regulator"}, {"减少畸形果", "regulator"},
		{"提高结铃率", "regulator"}, {"促进花粉萌发", "regulator"}, {"减少空壳", "regulator"},
		{"减少空荚现象", "regulator"}, {"预防苦痘病", "regulator"}, {"预防脐腐病", "regulator"}, {"预防干烧心", "regulator"},
		{"预防叶焦病", "regulator"}, {"预防黄叶病", "regulator"}, {"增强硬度", "regulator"},
		{"延长保鲜期", "regulator"}, {"增强养分吸收能力", "regulator"}, {"转色增甜", "regulator"},
		{"减少早衰现象", "regulator"}, {"增强细胞壁形成", "regulator"}, {"防止果实开裂", "regulator"},
		{"改善糖份运输", "regulator"}, {"提升果实甜度", "regulator"}, {"参与氮代谢", "regulator"},
		{"促进豆科作物根瘤固氮", "regulator"}, {"提高蛋白质合成", "regulator"},
	}

	tgtSQL := `INSERT OR IGNORE INTO targets (name, type, sort_order) VALUES (?, ?, ?)`
	for i, t := range targets {
		if _, err = tx.Exec(tgtSQL, t.name, t.typ, i); err != nil {
			return err
		}
	}

	return tx.Commit()
}
