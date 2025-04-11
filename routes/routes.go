package routes

import (
	//"intern_template_v1/controller"
	// "intern_template_v1/controller"
	admincontroller "intern_template_v1/controller/admin"
	admincontroller2 "intern_template_v1/controller/admin/user_management"
	authcontroller "intern_template_v1/controller/auth"

	// Usercontroller "intern_template_v1/controller/auth"
	all "intern_template_v1/controller/all"

	landlordcontroller "intern_template_v1/controller/landlord"
	landlordcontroller2 "intern_template_v1/controller/landlord/business_profile"

	landlordcontroller_inquiries "intern_template_v1/controller/landlord/inquries"

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
	go tenantscontroller.DeleteExpiredInquiries()
	//////////////////// Landlord //////////////////
	app.Post("/property/add", middleware.AuthMiddleware, landlordcontroller.CreateApartment)          //insert application for landlord apartment
	app.Get("/property/get", middleware.AuthMiddleware, landlordcontroller.FetchApartmentsByLandlord) //Property get by landlord

	app.Post("/create/businessname", middleware.AuthMiddleware, landlordcontroller2.UpdateBusinessName)             // insert business name
	app.Post("/create/businesspermit", middleware.AuthMiddleware, landlordcontroller2.SetUpdateBusinessPermitImage) //business permit

	app.Get("/tenants/inquiry/display", middleware.AuthMiddleware, landlordcontroller.FetchInquiriesByLandlord) // Fetch tenants inquiry
	app.Put("/landlord/inquiry/status", middleware.AuthMiddleware, landlordcontroller_inquiries.UpdateInquiryStatusByLandlord)

	app.Put("/update-inquiry-status/:uid", landlordcontroller.FetchInquiriesByLandlord) // Approve/Reject a users inquiry

	app.Delete("/apartment/delete/:id", middleware.AuthMiddleware, landlordcontroller.DeleteApartment) // landlord confirms rejected apartment

	//////////////////// Landlord //////////////////

	//////////////////// Admin //////////////////
	app.Post("/admin/register", admincontroller.RegisterAdmin) // register admin
	app.Post("/admin/login", admincontroller.LoginHandler)     // login admin

	app.Get("/display/users", admincontroller2.GetFilteredUserDetails) // Fetch all users can be filtered through name=John,accountname=artem&user_type=Landlord
	app.Put("/users/update", admincontroller2.UpdateUserDetails)       // Updating user values in the admin
	app.Delete("/admin/user/:uid", admincontroller2.SoftDeleteUser)    // Mark the account status as deleted

	app.Get("/admin/apartments/details", admincontroller2.GetApartmentDetails) //Get complete apartment details along with other data

	app.Put("/admin/promoting/account/:uid", admincontroller.UpdateUserType) //update user type tenant / land;lord
	app.Get("/apartments/pending", admincontroller.GetPendingApartments)     // Fetch unverified apartments
	app.Put("/apartments/verify/:id", admincontroller.VerifyApartment)       // Approve/Reject an apartment

	//////////////////// Admin //////////////////

	//	FOR ALL
	app.Post("/create/validid", middleware.AuthMiddleware, all.SetValidID)

	//////////////////// Tenant //////////////////
	app.Post("/create/inquiry", middleware.AuthMiddleware, tenantscontroller.CreateInquiry)
	app.Get("/fetchpending/inquiry", middleware.AuthMiddleware, tenantscontroller.FetchPendingInquiriesForTenant)
	app.Post("/tenant/delete-inquiry", middleware.AuthMiddleware, tenantscontroller.DeleteInquiryAfterViewingNotification)
	app.Get("/tenant/inquiries/count-status", middleware.AuthMiddleware, tenantscontroller.CountAcceptedOrRejectedInquiries)
	app.Get("/tenant/inquiries/get-notification", middleware.AuthMiddleware, tenantscontroller.GetApprovedOrRejectedInquiries)

	app.Get("/api/apartments/Approved", tenantscontroller.FetchApprovedApartmentsForTenant) //Display all the Approved apartment

	app.Post("/add/wishlist", middleware.AuthMiddleware, tenantscontroller.AddToWishlist)
	app.Get("/get/wishlist", middleware.AuthMiddleware, tenantscontroller.FetchwishlistForTenant)
	app.Delete("/wishlist/:apartment_id", middleware.AuthMiddleware, tenantscontroller.RemoveFromWishlist)

	//////////////////// Tenant //////////////////

	//////////////////// ALL //////////////////
	app.Put("/api/user/update-contact", middleware.AuthMiddleware, landlordcontroller2.UpdateContactInfo)

	//////////////////// ALL //////////////////

	//route for landlord verification
	app.Get("/user/pending", admincontroller.GetPendingUsers) // Fetch unverified users
	app.Put("/user/verify/:id", admincontroller.VerifyUsers)  // Approve/Reject a users

	//route for getting / displaying tenants inquiry

	//routes for tenants application inquiry and viewing approved dashboad
	// //app.Post("/api/inquiry/application", middleware.AuthMiddleware, tenantscontroller.CreateInquiry) //inquire in specific apartment
	app.Post("/firebase", authcontroller.VerifyFirebaseToken)
	//rountes for automatically deleting tenants inquiry
	//app.Get("/inquiries/cleanup", middleware.AuthMiddleware, tenantscontroller.NotifyPendingInquiries)



	//routes for uploading images and videos
	// app.Post("/apartment/:apartment_id/image", controller.UploadApartmentImageHandler)
	// app.Post("/upload/video", controller.UploadVideoHandler)


}


