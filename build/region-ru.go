package main

var regionRU = regionDef{
	Domain: "sh-development.ru",
	Lang:   "ru",
	Locale: "ru-RU",

	AIScan:     false,
	AIFoodLang: "Russian",

	Text: pageText{
		Login: "войти",
		About: "Считает калории и БЖУ по фото или описанию еды. Ведет учет калорий и отслеживает изменения веса.",

		TabFood:   "Еда",
		TabWeight: "Вес",
		TabInfo:   "Инфо",
		Apps:      "apps",
		Readme:    "readme",
		Logout:    "выйти",

		KcalToday: "ккал сегодня",
		ProtUnit:  "г белка",
		Manual:    "Вручную",
		AIPhoto:   "ИИ фото",
		AIText:    "ИИ текст",
		Favorites: "Избранное",
		TodayHead: "За сегодня",
		Kcal:      "ккал",

		Week:        "Неделя",
		Month:       "Месяц",
		Year:        "Год",
		NoData:      "Нет записей за этот период",
		WeightToday: "Вес сегодня",
		Kg:          "кг",
		Save:        "Сохранить",
		Date:        "Дата",

		Name:       "Имя",
		YourID:     "Ваш ID",
		Status:     "Статус",
		ScansToday: "Сканов сегодня",
		BuyPro:     "Купить PRO",
		ProText:    "Магазина пока нет — проект nom-nom на бета-тестировании. Если хочешь принять участие, свяжись со мной:",

		MealName: "Название",
		Grams:    "Граммы",
		Prot:     "Белки",
		Fat:      "Жиры",
		Carb:     "Углеводы",

		GoalsTitle: "Цели на день",
		GoalKcal:   "Калории, ккал",
		GoalProt:   "Белок, г",

		DescribeHint: "Опиши блюдо…  например: «тарелка овсянки с бананом и мёдом»",
		PointCamera:  "Наведи камеру на еду",
		Analyzing:    "Анализирую…",

		FavSearch:      "Блюдо",
		FavSearchLabel: "Поиск по избранному",

		OK: "Ок",

		JS: jsText{
			Error:         "Ошибка",
			NewDay:        "Новый день",
			NoMeals:       "Вы еще не ном-ном сегодня",
			GoalKcalRange: "Калории: от 100 до 20000",
			GoalProtRange: "Белок: от 10 до 1000 г",
			GoalsSaved:    "Цели сохранены",
			EmailCopied:   "Email скопирован",
			Today:         "Сегодня",
			Edit:          "Изменить",
			NoWeights:     "Записей пока нет",
			WeightPrompt:  "Вес за %s (кг)",
			EnterWeight:   "Введи вес в кг",
			WeightSaved:   "Вес сохранён",
			ScanSaved:     "Скан сохранён в приёмы за сегодня",
			DefaultMeal:   "Блюдо",
			MealDeleted:   "Приём удалён",
			NoFavorites:   "Избранного пока нет — отметь блюдо звездой ★.",
			NotFound:      "Ничего не найдено.",
			GramsUnit:     "г",
			PickFirst:     "Сначала выбери блюдо",
			ScanFailed:    "Упс, что-то пошло не так. Попробуй ещё раз позже.",
			NoCamera:      "Камера недоступна — выбери фото из галереи",
			DescribeMeal:  "Опиши блюдо",
			Unavailable:   "К сожалению, эта функция недоступна в вашем регионе",
		},
	},

	Server: serverText{
		DailyLimit:   "Дневной лимит сканов достигнут",
		NoFreeScans:  "Бесплатные сканы закончились",
		ServiceDown:  "Сервис временно недоступен",
		AITimeout:    "Не дождались ответа от ИИ",
		StatusTester: "Тестер · %d/день",
		StatusFree:   "AI сканов: %d",
	},
}
