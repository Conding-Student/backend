package routes

import (
	//"intern_template_v1/controller"
	admincontroller "intern_template_v1/controller/admin"
	authcontroller "intern_template_v1/controller/auth"

	// Usercontroller "intern_template_v1/controller/auth"
	all "intern_template_v1/controller/all"
	landlordcontroller "intern_template_v1/controller/landlord"
	tenantscontroller "intern_template_v1/controller/tenants"
	"intern_template_v1/middleware"

	"github.com/gofiber/fiber/v2"
	//"golang.org/x/crypto/nacl/auth"
)

func AppRoutes(app *fiber.App) {
	// SAMPLE ENDPOINT
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("RentXpert! go, go, go lang!")
	})

	// CREATE YOUR ENDPOINTS HERE
	//app.Get("/try", controller.SampleController2)
	//app.Get("/try1", controller.SampleController1)

	//app.Post("/create", controller.UserRegistration)
	//app.Post("/read", controller.ReadUser)
	//app.Post("/create", controller.CreateBook)
	//app.Get("/get/all/books", controller.GetAllBooks)
	//app.Get("/get/books/:id", controller.Getbook)
	//app.Put("/update/book/:id", controller.UpdateBook)
	//app.Post("/register/user", controller.RegisterUser)
	//app.Get("/get/user/:id", controller.GetUser)
	//app.Get("/get/all/user", controller.GetAllUsers)
	//testing
	// app.Post("/registertenant/account", Usercontroller.RegisterTenant)
	// app.Post("/registerlandlord/account", Usercontroller.RegisterLandlord)
	// app.Post("/loginuser/account", Usercontroller.LoginUser)

	//app.Post("/addrentallisting", landlordcontroller.CreateApartment)
	app.Post("/property/add", middleware.AuthMiddleware, landlordcontroller.CreateApartment)                    //insert application for landlord apartment
	app.Get("/property/get", middleware.AuthMiddleware, landlordcontroller.FetchApartmentsByLandlord)           //Property get by landlord
	app.Post("/create/businessname", middleware.AuthMiddleware, landlordcontroller.SetBusinessName)             // insert business name
	app.Post("/create/businesspermit", middleware.AuthMiddleware, landlordcontroller.SetBusinessPermitImage)    //business permit
	app.Get("/tenants/inquiry/display", middleware.AuthMiddleware, landlordcontroller.FetchInquiriesByLandlord) // Fetch tenants inquiry
	app.Put("/update-inquiry-status/:uid", landlordcontroller.UpdateInquiryStatusByLandlord)                    // Approve/Reject a users inquiry

	// Admin
	app.Post("/admin/register", admincontroller.RegisterAdmin) // register admin
	app.Post("/admin/login", admincontroller.LoginHandler)     // login admin
	app.Put("/admin/promoting/account/:uid", admincontroller.UpdateUserType)
	app.Get("/apartments/pending", admincontroller.GetPendingApartments) // Fetch unverified apartments
	app.Put("/apartments/verify/:id", admincontroller.VerifyApartment)   // Approve/Reject an apartment
	app.Delete("/apartment/:id/delete", admincontroller.ConfirmLandlord) // landlord confirms rejected apartment
	//	FOR ALL
	app.Post("/create/validid", middleware.AuthMiddleware, all.SetValidID)

	// Tenant
	app.Post("/create/inquiry", middleware.AuthMiddleware, tenantscontroller.CreateInquiry)
	app.Get("/api/apartments/Approved", tenantscontroller.FetchApprovedApartmentsForTenant) //Display all the Approved apartment
	app.Post("/add/wishlist", middleware.AuthMiddleware, tenantscontroller.AddToWishlist)
	app.Delete("/wishlist/:apartment_id", middleware.AuthMiddleware, tenantscontroller.RemoveFromWishlist)

	app.Get("/get/wishlist", middleware.AuthMiddleware, tenantscontroller.FetchwishlistForTenant)
	//app.Get("/add/wishlist", middleware.AuthMiddleware, tenantscontroller.di)
	// route for admin registration/login
	// route for admin registration/login

	//route for apartment verification

	//Landlord confirms "rejected" apartment info
	app.Delete("/apartment/:id/delete", admincontroller.ConfirmLandlord) // landlord confirms rejected apartment

	//route for landlord verification
	app.Get("/user/pending", admincontroller.GetPendingUsers) // Fetch unverified users
	app.Put("/user/verify/:id", admincontroller.VerifyUsers)  // Approve/Reject a users

	//route for getting / displaying tenants inquiry

	//routes for tenants application inquiry and viewing approved dashboad
	// //app.Post("/api/inquiry/application", middleware.AuthMiddleware, tenantscontroller.CreateInquiry) //inquire in specific apartment
	app.Post("/firebase", authcontroller.VerifyFirebaseToken)
	//rountes for automatically deleting tenants inquiry
	//app.Get("/inquiries/cleanup", middleware.AuthMiddleware, tenantscontroller.NotifyPendingInquiries)

}
