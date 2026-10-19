package main

var regionCOM = regionDef{
	Domain: "sh-development.com",
	Lang:   "en",
	Locale: "en-US",

	AIScan:     true,
	AIFoodLang: "English",

	Text: pageText{
		Login: "log in",
		About: "Counts calories and macros from a photo or a description of your food. Tracks your calories and your weight over time.",

		TabFood:   "Food",
		TabWeight: "Weight",
		TabInfo:   "Info",
		Apps:      "apps",
		Readme:    "readme",
		Logout:    "log out",

		KcalToday: "kcal today",
		ProtUnit:  "g protein",
		Manual:    "Manual",
		AIPhoto:   "AI photo",
		AIText:    "AI text",
		Favorites: "Favorites",
		TodayHead: "Today",
		Kcal:      "kcal",

		Week:        "Week",
		Month:       "Month",
		Year:        "Year",
		NoData:      "No entries for this period",
		WeightToday: "Weight today",
		Kg:          "kg",
		Save:        "Save",
		Date:        "Date",

		Name:       "Name",
		YourID:     "Your ID",
		Status:     "Status",
		ScansToday: "Scans today",
		BuyPro:     "Buy PRO",
		ProText:    "There's no shop yet — nom-nom is in beta testing. If you'd like to take part, get in touch:",

		MealName: "Name",
		Grams:    "Grams",
		Prot:     "Protein",
		Fat:      "Fat",
		Carb:     "Carbs",

		GoalsTitle: "Daily goals",
		GoalKcal:   "Calories, kcal",
		GoalProt:   "Protein, g",

		DescribeHint: "Describe the meal…  e.g. “a bowl of oatmeal with banana and honey”",
		PointCamera:  "Point the camera at your food",
		Analyzing:    "Analyzing…",

		FavSearch:      "Dish",
		FavSearchLabel: "Search favorites",

		OK: "Ok",

		JS: jsText{
			Error:         "Error",
			NewDay:        "New day",
			NoMeals:       "No nom-nom yet today",
			GoalKcalRange: "Calories: 100 to 20000",
			GoalProtRange: "Protein: 10 to 1000 g",
			GoalsSaved:    "Goals saved",
			EmailCopied:   "Email copied",
			Today:         "Today",
			Edit:          "Edit",
			NoWeights:     "No entries yet",
			WeightPrompt:  "Weight on %s (kg)",
			EnterWeight:   "Enter your weight in kg",
			WeightSaved:   "Weight saved",
			ScanSaved:     "Scan saved to today's meals",
			DefaultMeal:   "Meal",
			MealDeleted:   "Meal deleted",
			NoFavorites:   "No favorites yet — star a meal ★ to add one.",
			NotFound:      "Nothing found.",
			GramsUnit:     "g",
			PickFirst:     "Pick a meal first",
			ScanFailed:    "Oops, something went wrong. Please try again later.",
			NoCamera:      "Camera unavailable — pick a photo from the gallery",
			DescribeMeal:  "Describe the meal",
			Unavailable:   "Sorry, this feature is not available in your region",
		},
	},

	Server: serverText{
		DailyLimit:   "Daily scan limit reached",
		NoFreeScans:  "No free scans left",
		ServiceDown:  "Service temporarily unavailable",
		AITimeout:    "The AI took too long to answer",
		StatusTester: "Tester · %d/day",
		StatusFree:   "AI scans: %d",
	},
}
