package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// 数据库连接配置
const (
	DBHost     = "localhost"
	DBPort     = 5432
	DBUser     = "gaokao_user"
	DBPassword = "gaokao_password"
	DBName     = "gaokao_user_db"
)

// ExtendedMajor 结构体，对应extended-data-service.go中的数据
type ExtendedMajor struct {
	ID           int    `json:"id"`
	UniversityID int    `json:"university_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	IsActive     bool   `json:"is_active"`
	Duration     int    `json:"duration"`
	Degree       string `json:"degree"`
	Description  string `json:"description"`
}

// 连接数据库
func connectDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPassword, DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("成功连接到数据库!")
	return db, nil
}

// 检查majors表是否存在
func checkMajorsTable(db *sql.DB) error {
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = 'majors'
		)
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("检查majors表失败: %v", err)
	}

	if !tableExists {
		return fmt.Errorf("majors表不存在，请先运行数据库初始化脚本")
	}

	fmt.Println("✓ majors表存在")
	return nil
}

// 生成专业数据（从extended-data-service.go复制）
func generateExtendedMajors() []ExtendedMajor {
	return []ExtendedMajor{
		{ID: 1, UniversityID: 1, Code: "080901", Name: "计算机科学与技术", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具有良好的科学素养，系统地、较好地掌握计算机科学与技术包括计算机硬件、软件与应用的基本理论、基本知识和基本技能与方法的高级专门科学技术人才。"},
		{ID: 2, UniversityID: 1, Code: "080902", Name: "软件工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养适应计算机应用学科的发展，特别是软件产业的发展，具备计算机软件的基础理论、基本知识和基本技能，具有用软件工程的思想、方法和技术来分析、设计和实现计算机软件系统的能力的高级专门人才。"},
		{ID: 3, UniversityID: 1, Code: "080903", Name: "网络工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养掌握网络工程的基本理论与方法以及计算机技术和网络技术等方面的知识，能运用所学知识与技能去分析和解决相关的实际问题，可在信息产业以及其他国民经济部门从事各类网络系统和计算机通信系统研究、教学、设计、开发等工作的高级网络科技人才。"},
		{ID: 4, UniversityID: 1, Code: "080905", Name: "物联网工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养能够系统地掌握物联网的相关理论、方法和技能，具备通信技术、网络技术、传感技术等信息领域宽广的专业知识的高级工程技术人才。"},
		{ID: 5, UniversityID: 1, Code: "080906", Name: "数字媒体技术", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具有良好的科学素养以及美术修养、既懂技术又懂艺术、能利用计算机新的媒体设计工具进行艺术作品的设计和创作的复合型应用设计人才。"},
		{ID: 6, UniversityID: 2, Code: "120201", Name: "工商管理", Category: "管理学", IsActive: true, Duration: 4, Degree: "管理学学士", Description: "培养具备管理、经济、法律及企业管理方面的知识和能力，能在企、事业单位及政府部门从事管理以及教学、科研方面工作的工商管理学科高级专门人才。"},
		{ID: 7, UniversityID: 2, Code: "120202", Name: "市场营销", Category: "管理学", IsActive: true, Duration: 4, Degree: "管理学学士", Description: "培养具备管理、经济、法律、市场营销等方面的知识和能力，能在企、事业单位及政府部门从事市场营销与管理以及教学、科研方面工作的工商管理学科高级专门人才。"},
		{ID: 8, UniversityID: 2, Code: "120203", Name: "会计学", Category: "管理学", IsActive: true, Duration: 4, Degree: "管理学学士", Description: "培养具备管理、经济、法律和会计学等方面的知识和能力，能在企、事业单位及政府部门从事会计实务以及教学、科研方面工作的工商管理学科高级专门人才。"},
		{ID: 9, UniversityID: 2, Code: "120204", Name: "财务管理", Category: "管理学", IsActive: true, Duration: 4, Degree: "管理学学士", Description: "培养具备管理、经济、法律和理财、金融等方面的知识和能力，能在工商、金融企业、事业单位及政府部门从事财务、金融管理以及教学、科研方面工作的工商管理学科高级专门人才。"},
		{ID: 10, UniversityID: 2, Code: "120206", Name: "人力资源管理", Category: "管理学", IsActive: true, Duration: 4, Degree: "管理学学士", Description: "培养具备管理、经济、法律及人力资源管理等方面的知识和能力，能在事业单位及政府部门从事人力资源管理以及教学、科研方面工作的工商管理学科高级专门人才。"},
		{ID: 11, UniversityID: 3, Code: "020101", Name: "经济学", Category: "经济学", IsActive: true, Duration: 4, Degree: "经济学学士", Description: "培养具备比较扎实的马克思主义经济学理论基础，熟悉现代西方经济学理论，比较熟练地掌握现代经济分析方法，知识面较宽，具有向经济学相关领域扩展渗透的能力的专门人才。"},
		{ID: 12, UniversityID: 3, Code: "020201", Name: "财政学", Category: "经济学", IsActive: true, Duration: 4, Degree: "经济学学士", Description: "培养具备财政、税务等方面的理论知识和业务技能，能在财政、税务及其他经济管理部门和企业从事相关工作的高级专门人才。"},
		{ID: 13, UniversityID: 3, Code: "020301", Name: "金融学", Category: "经济学", IsActive: true, Duration: 4, Degree: "经济学学士", Description: "培养具有金融学理论知识及专业技能的专门人才。"},
		{ID: 14, UniversityID: 3, Code: "020401", Name: "国际经济与贸易", Category: "经济学", IsActive: true, Duration: 4, Degree: "经济学学士", Description: "培养的学生应较系统地掌握马克思主义经济学基本原理和国际经济、国际贸易的基本理论，掌握国际贸易的基本知识与基本技能。"},
		{ID: 15, UniversityID: 4, Code: "050101", Name: "汉语言文学", Category: "文学", IsActive: true, Duration: 4, Degree: "文学学士", Description: "培养具备文艺理论素养和系统的汉语言文学知识，能在新闻文艺出版部门、高校、科研机构和机关企事业单位从事文学评论、汉语言文学教学与研究工作，以及文化、宣传方面的实际工作的汉语言文学高级专门人才。"},
		{ID: 16, UniversityID: 4, Code: "050201", Name: "英语", Category: "文学", IsActive: true, Duration: 4, Degree: "文学学士", Description: "培养具有扎实的英语语言基础和比较广泛的科学文化知识，能在外事、经贸、文化、新闻出版、教育、科研、旅游等部门从事翻译、研究、教学、管理工作的英语高级专门人才。"},
		{ID: 17, UniversityID: 4, Code: "050207", Name: "日语", Category: "文学", IsActive: true, Duration: 4, Degree: "文学学士", Description: "培养具有扎实的相应语言基础比较广泛的科学文化知识，能在外事、经贸、文化、新闻出版、教育、科研、旅游等部门从事翻译、研究、教学、管理工作的相应语言高级专门人才。"},
		{ID: 18, UniversityID: 4, Code: "050301", Name: "新闻学", Category: "文学", IsActive: true, Duration: 4, Degree: "文学学士", Description: "培养具备系统的新闻理论知识与技能、宽广的文化与科学知识，熟悉我国新闻、宣传政策法规，能在新闻、出版与宣传部门从事编辑、记者与管理等工作的新闻学高级专门人才。"},
		{ID: 19, UniversityID: 5, Code: "070101", Name: "数学与应用数学", Category: "理学", IsActive: true, Duration: 4, Degree: "理学学士", Description: "培养掌握数学科学的基本理论与基本方法，具备运用数学知识、使用计算机解决实际问题的能力，受到科学研究的初步训练的高级专门人才。"},
		{ID: 20, UniversityID: 5, Code: "070201", Name: "物理学", Category: "理学", IsActive: true, Duration: 4, Degree: "理学学士", Description: "培养掌握物理学的基本理论与方法，具有良好的数学基础和实验技能，能在物理学或相关的科学技术领域中从事科研、教学、技术和相关的管理工作的高级专门人才。"},
		{ID: 21, UniversityID: 5, Code: "070301", Name: "化学", Category: "理学", IsActive: true, Duration: 4, Degree: "理学学士", Description: "培养具备化学的基础知识、基本理论和基本技能，能在化学及与化学相关的科学技术和其它领域从事科研、教学技术及相关管理工作的高级专门人才。"},
		{ID: 22, UniversityID: 5, Code: "071001", Name: "生物科学", Category: "理学", IsActive: true, Duration: 4, Degree: "理学学士", Description: "培养具备生物科学的基本理论、基本知识和较强的实验技能，能在科研机构、高等学校及企事业单位等从事科学研究、教学工作及管理工作的生物科学高级专门人才。"},
		{ID: 23, UniversityID: 6, Code: "100201", Name: "临床医学", Category: "医学", IsActive: true, Duration: 5, Degree: "医学学士", Description: "培养具备基础医学、临床医学的基本理论和医疗预防的基本技能；能在医疗卫生单位、医学科研等部门从事医疗及预防、医学科研等方面工作的医学高级专门人才。"},
		{ID: 24, UniversityID: 6, Code: "100301", Name: "口腔医学", Category: "医学", IsActive: true, Duration: 5, Degree: "医学学士", Description: "培养具备医学基础理论和临床医学知识，掌握口腔医学的基本理论和临床操作技能，能在医疗卫生机构从事口腔常见病、多发病的诊治、修复和与预防工作的医学高级专门人才。"},
		{ID: 25, UniversityID: 6, Code: "100401", Name: "预防医学", Category: "医学", IsActive: true, Duration: 5, Degree: "医学学士", Description: "培养具备基础医学、临床医学和预防医学的基本理论知识和防疫工作的基本能力，能在疾病控制中心、环境卫生监测、食品卫生监测等机构从事预防医学工作的高级专门人才。"},
		{ID: 26, UniversityID: 6, Code: "101101", Name: "护理学", Category: "医学", IsActive: true, Duration: 4, Degree: "理学学士", Description: "培养具备人文社会科学、医学、预防保健的基本知识及护理学的基本理论知识和技能，能在护理领域内从事临床护理、预防保健、护理管理、护理教学和护理科研的高级专门人才。"},
		{ID: 27, UniversityID: 7, Code: "040101", Name: "教育学", Category: "教育学", IsActive: true, Duration: 4, Degree: "教育学学士", Description: "培养具有良好思想道德品质、较高教育理论素养和较强教育实际工作能力的中、高等师范院校师资、中小学校教育科研人员、教育科学研究单位研究人员、各级教育行政管理人员和其他教育工作者。"},
		{ID: 28, UniversityID: 7, Code: "040106", Name: "学前教育", Category: "教育学", IsActive: true, Duration: 4, Degree: "教育学学士", Description: "培养具备学前教育专业知识，能在托幼机构从事保教和研究工作的教师学前教育行政人员以及其他有关机构的教学、研究人才。"},
		{ID: 29, UniversityID: 7, Code: "040107", Name: "小学教育", Category: "教育学", IsActive: true, Duration: 4, Degree: "教育学学士", Description: "培养德、智、体全面发展的，具有较高教育理论素养和较强教育实际工作能力（语、数、英）小学教师及教育科研、各级教育行政管理人员和其他教育工作者。"},
		{ID: 30, UniversityID: 7, Code: "040201", Name: "体育教育", Category: "教育学", IsActive: true, Duration: 4, Degree: "教育学学士", Description: "培养具备现代教育理念，能够较好地掌握体育教育的基本理论、知识和技能，掌握学校体育教育工作规律，具有较强的实践能力，在全面发展的基础上有所专长的体育专门人才。"},
		{ID: 31, UniversityID: 8, Code: "030101", Name: "法学", Category: "法学", IsActive: true, Duration: 4, Degree: "法学学士", Description: "培养系统掌握法学知识，熟悉我国法律和党的相关政策，能在国家机关、企事业单位和社会团体、特别是能在立法机关、行政机关、检察机关、审判机关、仲裁机构和法律服务机构从事法律工作的高级专门人才。"},
		{ID: 32, UniversityID: 8, Code: "030201", Name: "政治学与行政学", Category: "法学", IsActive: true, Duration: 4, Degree: "法学学士", Description: "培养具有一定马克思主义理论素养和政治学、行政学方面的基本理论和专门知识，能在党政机关、新闻出版机构、企事业和社会团体等单位从事教学科研、行政管理等方面工作的政治学和行政学高级专门人才。"},
		{ID: 33, UniversityID: 8, Code: "030302", Name: "社会工作", Category: "法学", IsActive: true, Duration: 4, Degree: "法学学士", Description: "培养具有基本的社会工作理论和知识，较熟练的社会调查研究技能和社会工作能力，能在民政、劳动、杜会保障和卫生部门，及工会、青年、妇女等社会组织及其他社会福利、服务和公益团体等机构从事社会保障、社会政策研究、社会行政管理、社区发展与管理、社会服务、评估与操作等工作的高级专门人才。"},
		{ID: 34, UniversityID: 9, Code: "080201", Name: "机械工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备机械设计、制造、机电工程及自动化基础知识与应用能力，能在工业生产第一线从事机械制造领域内的设计制造、科技开发、应用研究、运行管理和经营销售等方面工作的高级工程技术人才。"},
		{ID: 35, UniversityID: 9, Code: "080202", Name: "机械设计制造及其自动化", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备机械设计制造基础知识与应用能力，能在工业主产第一线从事机械制造领域内的设计制造、科技开发、应用研究、运行管理和经营销售等方面工作的高级工程技术人才。"},
		{ID: 36, UniversityID: 9, Code: "080301", Name: "测控技术与仪器", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备精密仪器设计制造以及测量与控制方面基础知识与应用能力，能在国民经济各部门从事测量与控制领域内有关技术、仪器与系统的设计制造、科技开发、应用研究、运行管理等方面的高级工程技术人才。"},
		{ID: 37, UniversityID: 9, Code: "080401", Name: "材料科学与工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备包括金属材料、无机非金属材料、高分子材料等材料领域的科学与工程方面较宽的基础知识，能在各种材料的制备、加工成型、材料结构与性能等领域从事科学研究与教学、技术开发、工艺和设备设计、技术改造及经营管理等方面工作的高级工程技术人才。"},
		{ID: 38, UniversityID: 10, Code: "080601", Name: "电气工程及其自动化", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养能够从事与电气工程有关的系统运行、自动控制、电力电子技术、信息处理、试验分析、研制开发、经济管理以及电子与计算机技术应用等领域工作的宽口径复合型高级工程技术人才。"},
		{ID: 39, UniversityID: 10, Code: "080701", Name: "电子信息工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备电子技术和信息系统的基础知识，能从事各类电子设备和信息系统的研究、设计、制造、应用和开发的高等工程技术人才。"},
		{ID: 40, UniversityID: 10, Code: "080703", Name: "通信工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备通信技术、通信系统和通信网等方面的知识，能在通信领域中从事研究、设计、制造、运营及在国民经济各部门和国防工业中从事开发、应用通信技术与设备的高级工程技术人才。"},
		{ID: 41, UniversityID: 10, Code: "080801", Name: "自动化", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养的学生要具备电工技术、电子技术、控制理论、自动检测与仪表、信息处理、系统工程、计算机技术与应用和网络技术等较宽广领域的工程技术基础和一定的专业知识的高级工程技术人才。"},
		{ID: 42, UniversityID: 11, Code: "081001", Name: "土木工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养掌握各类土木工程学科的基本理论和基本知识，能在房屋建筑、地下建筑、道路、隧道、桥梁建筑、水电站、港口及近海结构与设施工作的人员。"},
		{ID: 43, UniversityID: 11, Code: "081002", Name: "建筑环境与能源应用工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备室内环境设备系统及建筑公共设施系统的设计、安装调试、运行管理及国民经济各部门所需的特殊环境的研究开发的基础理论知识及能力的高级工程技术人才。"},
		{ID: 44, UniversityID: 11, Code: "081003", Name: "给排水科学与工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备城市给水工程、排水工程、建筑给水排水工程、工业给水排水工程、水污染控制规划和水资源保护等方面的知识，能在政府部门、规划部门、经济管理部门、环保部门、设计单位、工矿企业、科研单位、大、中专院校等从事规划、设计、施工、管理、教育和研究开发方面工作的给水排水工程学科的高级工程技术人才。"},
		{ID: 45, UniversityID: 11, Code: "082801", Name: "建筑学", Category: "工学", IsActive: true, Duration: 5, Degree: "建筑学学士", Description: "培养具备建筑设计、城市设计、室内设计等方面的知识，能在设计部门从事设计工作，并具有多种职业适应能力的通用型、复合型高级工程技术人才。"},
		{ID: 46, UniversityID: 12, Code: "082502", Name: "环境工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备城市和城镇水、气、声、固体废物等污染防治和给排水工程、污染控制规划和水资源保护等方面的知识，能在政府部门、规划部门、经济管理部门、环保部门、设计单位、工矿企业、科研单位、学校等从事规划、设计、施工、管理、教育和研究开发方面工作的环境工程学科高级工程技术人才。"},
		{ID: 47, UniversityID: 12, Code: "083001", Name: "生物工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养掌握生物技术及其产业化的科学原理、工艺技术过程和工程设计等基础理论，基本技能，能在生物技术与工程领域从事设计生产管理和新技术研究、新产品开发的工程技术人才。"},
		{ID: 48, UniversityID: 12, Code: "082601", Name: "生物医学工程", Category: "工学", IsActive: true, Duration: 4, Degree: "工学学士", Description: "培养具备生命科学、电子技术、计算机技术及信息科学有关的基础理论知识以及医学与工程技术相结合的科学研究能力，能在生物医学工程领域、医学仪器以及其它电子技术、计算机技术、信息产业等部门从事研究、开发、教学及管理的高级工程技术人才。"},
		{ID: 49, UniversityID: 13, Code: "090101", Name: "农学", Category: "农学", IsActive: true, Duration: 4, Degree: "农学学士", Description: "培养具备作物生产、作物遗传育种以及种子生产与经营管理等方面的基本理论、基本知识和基本技能，能在农业及其它相关的部门或单位从事与农学有关的技术与设计、推广与开发、经营与管理、教学与科研等工作的高级科学技术人才。"},
		{ID: 50, UniversityID: 13, Code: "090102", Name: "园艺", Category: "农学", IsActive: true, Duration: 4, Degree: "农学学士", Description: "培养具备生物学和园艺学的基本理论、基本知识和基本技能，能在农业、商贸、园林管理等领域和部门从事与园艺科学有关的技术与设计、推广与开发、经营与管理、教学与科研等工作的高级科学技术人才。"},
		{ID: 51, UniversityID: 13, Code: "090201", Name: "农业资源与环境", Category: "农学", IsActive: true, Duration: 4, Degree: "农学学士", Description: "培养具备农业资源与环境方面的基本理论、基本知识和基本技能，能在农业、土地、环保、农资等部门或单位从事农业资源管理及利用、农业环境保护、生态农业、资源遥感与信息技术的教学、科研、管理等工作的高级科学技术人才。"},
		{ID: 52, UniversityID: 13, Code: "090401", Name: "动物医学", Category: "农学", IsActive: true, Duration: 5, Degree: "农学学士", Description: "培养具备动物医学方面的基本理论、基本知识和基本技能，能在兽医业务部门、动物生产单位及有关部门从事兽医、防疫检疫、教学、科学研究等工作的高级科学技术人才。"},
		{ID: 53, UniversityID: 14, Code: "130201", Name: "音乐表演", Category: "艺术学", IsActive: true, Duration: 4, Degree: "艺术学学士", Description: "培养具有一定的马克思主义基本理论素养，并具备音乐表演方面的能力，能在专业文艺团体、艺术院校等相关部门、机构从事表演、教学及研究工作的高级专门人才。"},
		{ID: 54, UniversityID: 14, Code: "130202", Name: "音乐学", Category: "艺术学", IsActive: true, Duration: 4, Degree: "艺术学学士", Description: "培养具有一定的马克思主义基本理论素养和系统的专业基本知识，具备一定音乐实践技能和教学能力，能在高、中等专业或普通院校、社会文艺团体、艺术研究单位和文化机关、出版及广播、影视部门从事教学、研究、编辑、评论、管理等方面工作的高级专门人才。"},
		{ID: 55, UniversityID: 14, Code: "130401", Name: "美术学", Category: "艺术学", IsActive: true, Duration: 4, Degree: "艺术学学士", Description: "培养具备美术学的基本理论、基本知识和基本技能，能够在高等和中等学校进行美术教学和教学研究的教师、教学研究人员和其他教育工作者。"},
		{ID: 56, UniversityID: 14, Code: "130502", Name: "视觉传达设计", Category: "艺术学", IsActive: true, Duration: 4, Degree: "艺术学学士", Description: "培养具有国际设计文化视野、中国设计文化特色、适合于创新时代需求，集传统平面（印刷）媒体和现代数字媒体，在专业设计领域、企业、传播机构、大企业市场部门、中等院校、研究单位从事视觉传播方面的设计、教学、研究和管理工作的专门人才。"},
		{ID: 57, UniversityID: 15, Code: "120401", Name: "公共事业管理", Category: "管理学", IsActive: true, Duration: 4, Degree: "管理学学士", Description: "培养具备现代管理理论、技术与方法等方面的知识以及应用这些知识的能力，能在文教、体育、卫生、环保、社会保险等公共事业单位行政管理部门从事管理工作的高级专门人才。"},
	}
}

// 导入专业数据
func importMajors(db *sql.DB) error {
	majors := generateExtendedMajors()

	// 准备插入语句 - 匹配实际的majors表结构
	stmt, err := db.Prepare(`
		INSERT INTO majors (university_id, code, name, category, discipline, degree_type, duration, description) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	// 首先获取一个大学ID作为默认值
	var defaultUniversityID string
	err = db.QueryRow("SELECT id FROM universities LIMIT 1").Scan(&defaultUniversityID)
	if err != nil {
		return fmt.Errorf("获取默认大学ID失败: %v", err)
	}

	successCount := 0
	errorCount := 0

	for _, major := range majors {
		// 执行插入
		_, err := stmt.Exec(
			defaultUniversityID, // 使用默认大学ID
			major.Code,
			major.Name,
			major.Category,
			major.Category, // discipline 使用 category 的值
			major.Degree,
			major.Duration,
			major.Description,
		)

		if err != nil {
			fmt.Printf("插入专业失败 %s: %v\n", major.Name, err)
			errorCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("专业数据导入完成: 成功 %d 条，失败 %d 条\n", successCount, errorCount)
	return nil
}

// 验证数据
func verifyMajorData(db *sql.DB) error {
	// 统计总数
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM majors").Scan(&totalCount)
	if err != nil {
		return fmt.Errorf("查询专业总数失败: %v", err)
	}
	fmt.Printf("数据库中共有 %d 个专业\n", totalCount)

	// 按学科门类统计
	rows, err := db.Query(`
		SELECT category, COUNT(*) as count 
		FROM majors 
		WHERE category IS NOT NULL AND category != '' 
		GROUP BY category 
		ORDER BY count DESC
	`)
	if err != nil {
		return fmt.Errorf("查询专业分类统计失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n按学科门类统计:")
	for rows.Next() {
		var category string
		var count int
		err := rows.Scan(&category, &count)
		if err != nil {
			return fmt.Errorf("扫描分类统计结果失败: %v", err)
		}
		fmt.Printf("  %s: %d 个专业\n", category, count)
	}

	// 检查是否有重复的专业代码
	var duplicateCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT code FROM majors 
			GROUP BY code 
			HAVING COUNT(*) > 1
		) as duplicates
	`).Scan(&duplicateCount)
	if err != nil {
		return fmt.Errorf("检查重复专业代码失败: %v", err)
	}

	if duplicateCount > 0 {
		fmt.Printf("\n警告: 发现 %d 个重复的专业代码\n", duplicateCount)
	} else {
		fmt.Println("\n✓ 没有发现重复的专业代码")
	}

	return nil
}

func main() {
	fmt.Println("开始导入专业数据...")

	// 连接数据库
	db, err := connectDB()
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 检查表是否存在
	err = checkMajorsTable(db)
	if err != nil {
		log.Fatal("检查表失败:", err)
	}

	// 导入专业数据
	err = importMajors(db)
	if err != nil {
		log.Fatal("导入专业数据失败:", err)
	}

	// 验证数据
	err = verifyMajorData(db)
	if err != nil {
		log.Fatal("验证数据失败:", err)
	}

	fmt.Println("\n专业数据导入完成!")
}