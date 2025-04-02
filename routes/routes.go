package routes

import (
	//"intern_template_v1/controller"
	controller "intern_template_v1/controller/Admin"
	Usercontroller "intern_template_v1/controller/auth"
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
	app.Post("/registertenant/account", Usercontroller.RegisterTenant)
	app.Post("/registerlandlord/account", Usercontroller.RegisterLandlord)
	app.Post("/loginuser/account", Usercontroller.LoginUser)

	//app.Post("/addrentallisting", landlordcontroller.CreateApartment)
	app.Post("/property/add", middleware.AuthMiddleware, landlordcontroller.CreateApartment)
	app.Get("/property/get", middleware.AuthMiddleware, landlordcontroller.FetchApartmentsByLandlord)

	// route for admin registration/login
	app.Post("/admin/register", controller.RegisterAdmin) // register admin
	app.Post("/admin/login", controller.LoginHandler)     // login admin

	//route for apartment verification
	app.Get("/apartments/pending", controller.GetPendingApartments) // Fetch unverified apartments
	app.Put("/apartments/verify/:id", controller.VerifyApartment)   // Approve/Reject an apartment

	//Landlord confirms "rejected" apartment info
	app.Delete("/apartment/:id/delete", controller.ConfirmLandlord) // landlord confirms rejected apartment

	//route for landlord verification
	app.Get("/user/pending", controller.GetPendingUsers) // Fetch unverified users
	app.Put("/user/verify/:id", controller.VerifyUsers)  // Approve/Reject a users

	//route for getting / displaying tenants inquiry
	app.Get("/tenants/inquiry", middleware.AuthMiddleware, landlordcontroller.FetchInquiriesByLandlord) // Fetch tenants inquiry
	app.Put("/tenants/inquiry/update/:id", landlordcontroller.UpdateInquiryStatus)                      // Approve/Reject a users

	//routes for tenants application inquiry and viewing approved dashboad
	app.Get("/api/apartments/Approved", tenantscontroller.FetchApprovedApartments)                   //Display all the Approved apartment
	app.Post("/api/inquiry/application", middleware.AuthMiddleware, tenantscontroller.CreateInquiry) //inquire in specific apartment

	//rountes for automatically deleting tenants inquiry
	app.Get("/inquiries/cleanup", middleware.AuthMiddleware, tenantscontroller.NotifyPendingInquiries)

}
