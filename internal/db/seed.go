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
		{2, 0, "杀虫"},
		{3, 0, "杀菌"},
		{4, 0, "杀螨剂"},
		{5, 0, "叶面肥"},
		{6, 0, "植物生长调节剂"},
		// 除草 二级分类
		{10, 1, "苗前除草剂"},
		{11, 1, "苗后除草剂"},
		{12, 1, "茎叶除草剂"},
		{13, 1, "除草联合套餐"},
		// 杀虫 二级分类
		{20, 2, "普通杀虫"},
		{21, 2, "种子处理"},
		// 杀菌 二级分类（暂无，使用一级 id=3）
		// 植物生长调节剂 二级（暂无）
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
		// 农作物
		"大豆", "小麦", "玉米", "水稻", "棉花", "花生", "甘蔗", "油菜",
		"高粱", "谷子", "向日葵", "甜菜", "大蒜", "韭菜",
		// 蔬菜
		"辣椒", "番茄", "黄瓜", "茄子", "白菜", "甘蓝", "花椰菜",
		"芹菜", "菠菜", "生菜", "萝卜", "胡萝卜", "土豆", "洋葱",
		"四季豆", "豇豆", "豌豆", "蚕豆", "韭黄", "莴苣", "青椒",
		"南瓜", "冬瓜", "丝瓜", "苦瓜", "西葫芦", "茴香", "香菜",
		"西红柿", "樱桃番茄",
		// 果树/经济作物
		"苹果", "梨", "桃", "杏", "李", "樱桃", "葡萄", "草莓",
		"柑橘", "柠檬", "荔枝", "龙眼", "芒果", "香蕉", "菠萝",
		"西瓜", "甜瓜", "哈密瓜", "山药", "魔芋",
		// 中药材/特种作物
		"山药", "党参", "黄芪", "白术", "枸杞", "丹参",
		// 除草场所（用于除草类产品）
		"果园", "茶园", "桑园", "公路", "铁路沿线", "荒山荒地",
		"绿化园林", "工业园区", "草坪", "高尔夫球场",
	}

	cropSQL := `INSERT OR IGNORE INTO crops (name, sort_order) VALUES (?, ?)`
	for i, name := range crops {
		if _, err = tx.Exec(cropSQL, name, i); err != nil {
			return err
		}
	}

	// ──────────────────────────────────────────────────────────────────────────
	// 3. 防治对象数据（targets）
	//    按 type 分三类：weed(杂草) / pest(害虫) / disease(病害)
	// ──────────────────────────────────────────────────────────────────────────
	type targetSeed struct {
		name     string
		typ      string
	}
	targets := []targetSeed{
		// 杂草 (weed)
		{"稗草", "weed"},
		{"马唐", "weed"},
		{"牛筋草", "weed"},
		{"狗尾草", "weed"},
		{"野燕麦", "weed"},
		{"看麦娘", "weed"},
		{"千金子", "weed"},
		{"藜", "weed"},
		{"苋", "weed"},
		{"反枝苋", "weed"},
		{"铁苋菜", "weed"},
		{"龙葵", "weed"},
		{"苘麻", "weed"},
		{"猪殃殃", "weed"},
		{"荠菜", "weed"},
		{"播娘蒿", "weed"},
		{"泽漆", "weed"},
		{"小飞蓬", "weed"},
		{"水花生", "weed"},
		{"水竹叶", "weed"},
		{"空心莲子草", "weed"},
		{"鳢肠", "weed"},
		{"鸭趾草", "weed"},
		{"香附子", "weed"},
		{"芦苇", "weed"},
		{"白茅", "weed"},
		{"狗牙根", "weed"},
		{"马齿苋", "weed"},
		{"繁缕", "weed"},
		{"问荆", "weed"},
		{"田旋花", "weed"},
		{"扁蓄", "weed"},
		{"蒲公英", "weed"},
		{"车前草", "weed"},
		{"婆婆纳", "weed"},
		{"野老鹳草", "weed"},
		{"刺儿菜", "weed"},
		{"大蓟", "weed"},
		{"苍耳", "weed"},
		{"一年蓬", "weed"},
		{"粘毛卷耳", "weed"},
		{"野高粱", "weed"},
		{"碎米莎草", "weed"},
		{"异型莎草", "weed"},
		{"牛毛毡", "weed"},
		{"眼子菜", "weed"},
		{"矮慈姑", "weed"},
		{"野慈姑", "weed"},
		{"扁秆藨草", "weed"},
		{"萤蔺", "weed"},
		{"水莎草", "weed"},
		{"三棱草", "weed"},
		{"雨久花", "weed"},
		{"浮萍", "weed"},
		{"节节草", "weed"},
		{"双穗雀稗", "weed"},
		{"通泉草", "weed"},
		{"遏蓝菜", "weed"},
		{"飞廉", "weed"},
		{"酸模", "weed"},
		{"曼陀罗", "weed"},
		{"龙葵", "weed"},
		{"燕麦草", "weed"},
		{"虎尾草", "weed"},
		{"臂形草", "weed"},
		{"早熟禾", "weed"},
		{"猫尾草", "weed"},
		{"鹅观草", "weed"},
		{"雀麦", "weed"},
		{"棒头草", "weed"},
		{"沙草", "weed"},
		{"黄花苜蓿", "weed"},
		{"野苜蓿", "weed"},
		{"苦荬菜", "weed"},

		// 害虫 (pest)
		{"蚜虫", "pest"},
		{"红蜘蛛", "pest"},
		{"白粉虱", "pest"},
		{"烟粉虱", "pest"},
		{"叶螨", "pest"},
		{"跗线螨", "pest"},
		{"茶黄螨", "pest"},
		{"蓟马", "pest"},
		{"灰飞虱", "pest"},
		{"白背飞虱", "pest"},
		{"褐飞虱", "pest"},
		{"稻纵卷叶螟", "pest"},
		{"二化螟", "pest"},
		{"三化螟", "pest"},
		{"稻苞虫", "pest"},
		{"水稻潜叶蝇", "pest"},
		{"潜叶蛾", "pest"},
		{"钻蛀性害虫", "pest"},
		{"玉米螟", "pest"},
		{"棉铃虫", "pest"},
		{"烟青虫", "pest"},
		{"斜纹夜蛾", "pest"},
		{"甜菜夜蛾", "pest"},
		{"小菜蛾", "pest"},
		{"菜青虫", "pest"},
		{"豆荚螟", "pest"},
		{"豆天蛾", "pest"},
		{"绿盲蝽", "pest"},
		{"美洲斑潜蝇", "pest"},
		{"南美斑潜蝇", "pest"},
		{"金纹细蛾", "pest"},
		{"卷叶蛾", "pest"},
		{"叶蝉", "pest"},
		{"梨木虱", "pest"},
		{"介壳虫", "pest"},
		{"刺蛾", "pest"},
		{"毒蛾", "pest"},
		{"尺蠖", "pest"},
		{"天幕毛虫", "pest"},
		{"舞毒蛾", "pest"},
		{"美国白蛾", "pest"},
		{"木橑尺蠖", "pest"},
		{"桃蛀螟", "pest"},
		{"桃小食心虫", "pest"},
		{"梨小食心虫", "pest"},
		{"桑天牛", "pest"},
		{"星天牛", "pest"},
		{"大豆食心虫", "pest"},
		{"豆杆黑潜蝇", "pest"},
		{"黄曲条跳甲", "pest"},
		{"蛴螬", "pest"},
		{"地老虎", "pest"},
		{"金针虫", "pest"},
		{"韭蛆", "pest"},
		{"根结线虫", "pest"},
		{"甘薯象甲", "pest"},

		// 病害 (disease)
		{"白粉病", "disease"},
		{"锈病", "disease"},
		{"霜霉病", "disease"},
		{"灰霉病", "disease"},
		{"炭疽病", "disease"},
		{"枯萎病", "disease"},
		{"立枯病", "disease"},
		{"猝倒病", "disease"},
		{"茎腐病", "disease"},
		{"根腐病", "disease"},
		{"褐斑病", "disease"},
		{"斑点落叶病", "disease"},
		{"轮纹病", "disease"},
		{"纹枯病", "disease"},
		{"稻瘟病", "disease"},
		{"恶苗病", "disease"},
		{"稻曲病", "disease"},
		{"苗枯病", "disease"},
		{"叶枯病", "disease"},
		{"条纹叶枯病", "disease"},
		{"黑穗病", "disease"},
		{"全蚀病", "disease"},
		{"赤霉病", "disease"},
		{"散黑穗病", "disease"},
		{"腥黑穗病", "disease"},
		{"大斑病", "disease"},
		{"小斑病", "disease"},
		{"灰斑病", "disease"},
		{"细菌性角斑病", "disease"},
		{"溃疡病", "disease"},
		{"疮痂病", "disease"},
		{"软腐病", "disease"},
		{"黄萎病", "disease"},
		{"青枯病", "disease"},
		{"疫病", "disease"},
		{"晚疫病", "disease"},
		{"早疫病", "disease"},
		{"病毒病", "disease"},
		{"花叶病", "disease"},
		{"蔓枯病", "disease"},
		{"菌核病", "disease"},
		{"黑斑病", "disease"},
		{"叶霉病", "disease"},
		{"煤污病", "disease"},
	}

	tgtSQL := `INSERT OR IGNORE INTO targets (name, type, sort_order) VALUES (?, ?, ?)`
	for i, t := range targets {
		if _, err = tx.Exec(tgtSQL, t.name, t.typ, i); err != nil {
			return err
		}
	}

	return tx.Commit()
}
