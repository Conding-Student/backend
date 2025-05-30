package routes

import (
	//"intern_template_v1/controller"
	// "intern_template_v1/controller"
	"log"
	"time"

	"github.com/Conding-Student/backend/config"
	admincontroller "github.com/Conding-Student/backend/controller/Admin"
	admincontroller3 "github.com/Conding-Student/backend/controller/Admin/apartment_management"
	admincontroller4 "github.com/Conding-Student/backend/controller/Admin/chart"
	admincontroller2 "github.com/Conding-Student/backend/controller/Admin/user_management"
	controller "github.com/Conding-Student/backend/controller/tenants"
	"github.com/Conding-Student/backend/handlers"

	authcontroller "github.com/Conding-Student/backend/controller/auth"

	// Usercontroller "intern_template_v1/controller/auth"
	all "github.com/Conding-Student/backend/controller/all"

	landlordcontroller "github.com/Conding-Student/backend/controller/landlord"
	landlordcontroller2 "github.com/Conding-Student/backend/controller/landlord/business_profile"

	landlordcontroller_inquiries "github.com/Conding-Student/backend/controller/landlord/inquries"

	tenantscontroller "github.com/Conding-Student/backend/controller/tenants"
	"github.com/Conding-Student/backend/middleware"

	"github.com/gofiber/fiber/v2"
	//"golang.org/x/crypto/nacl/auth"
)

func AppRoutes(app *fiber.App) {
	// SAMPLE ENDPOINT
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("RentXpert! go, go, go lang!")
	})
	//go tenantscontroller.DeleteExpiredInquiries()
	go landlordcontroller.ManageApartmentExpirations()
	go landlordcontroller.ManageExpiredDeletions()

	//////////////////// Landlord //////////////////

	/////////////////// PUT ////////////////////////
	app.Put("/apartments/:id/media", middleware.AuthMiddleware, landlordcontroller.UpdateApartmentMedia) // Adding images and videos
	app.Put("/landlord/apartmentupdate/:id", middleware.AuthMiddleware, landlordcontroller.UpdateApartment)
	app.Put("/landlord/apartments/updateavailability/:id", middleware.AuthMiddleware, landlordcontroller.UpdateApartmentAvailability) // Update the apartment details
	app.Put("/landlord/inquiry/status", middleware.AuthMiddleware, landlordcontroller_inquiries.UpdateInquiryStatusByLandlord)
	app.Put("/update-inquiry-status/:uid", landlordcontroller.FetchInquiriesByLandlord) // Approve/Reject a users inquiry

	/////////////////// POST ////////////////////////
	app.Post("/property/add", middleware.AuthMiddleware, landlordcontroller.CreateApartment) //insert application for landlord apartment
	//app.Post("/create/businessname", middleware.AuthMiddleware, landlordcontroller2.UpdateBusinessName)             // insert business name
	//app.Post("/create/businesspermit", middleware.AuthMiddleware, landlordcontroller2.SetUpdateBusinessPermitImage) //business permit

	app.Post("/admin/login", admincontroller.LoginHandler)
	app.Post("/bealandlord", middleware.AuthMiddleware, landlordcontroller.RegisterLandlord)                    //business permit
	app.Get("/property/get", middleware.AuthMiddleware, landlordcontroller.FetchApartmentsByLandlord)           //Property get by landlord
	app.Get("/tenants/inquiry/display", middleware.AuthMiddleware, landlordcontroller.FetchInquiriesByLandlord) // Fetch tenants inquiry

	/////////////////// DELETE ////////////////////////
	app.Delete("/apartment/delete/:id", middleware.AuthMiddleware, landlordcontroller.DeleteApartment)       // landlord confirms rejected apartment
	app.Delete("/apartment/deleteany/:id", middleware.AuthMiddleware, landlordcontroller.DeleteApartmentAny) // landlord delete any apartment

	//////////////////// Landlord //////////////////

	//////////////////// Admin //////////////////

	//////////////////// PUT //////////////////
	app.Put("/users/update", admincontroller2.UpdateUserDetails)                  // Updating user values in the admin
	app.Put("/admin/update-profile", admincontroller.UpdateAdminProfile)          // updating admin email or password
	app.Put("/admin/apartments/update/:id", admincontroller.UpdateApartmentInfo)  // Update the apartment details
	app.Put("/admin/promoting/account/:uid", admincontroller.UpdateUserType)      //update user type tenant / landlord
	app.Put("/admin/verifying/validid/:uid", admincontroller.UpdateAccountStatus) //update account status tenant / landlord
	app.Put("/apartments/verify/:id", admincontroller.VerifyApartment)            // Approve/Reject an apartment
	app.Put("/user/verify/:id", admincontroller.VerifyUsers)                      // Approve/Reject a users

	//////////////////// POST //////////////////
	app.Post("/admin/register", admincontroller.RegisterAdmin)                               // register admin
	app.Post("/admin/login", admincontroller.LoginHandler)                                   // login admin //password: yourSecurePassword123
	app.Put("/accept/landlordrequest/:uid", landlordcontroller2.VerifyLandlordUsingAdmin)    // aacepting landlord request
	app.Post("/rejecting/landlordrequest/:uid", landlordcontroller2.RejectLandlordRequest)   // rejecting landlord request
	app.Post("/rejecting/landlordApartment/:id", landlordcontroller2.RejectApartmentRequest) // rejecting landlord request
	app.Post("/firebase/login", authcontroller.VerifyFirebaseTokenAdmin)

	//////////////////// GET //////////////////
	app.Get("/adminuserinfo/search", admincontroller2.GetFilteredUserDetailspart2)
	app.Get("/api/stats/users-by-year", admincontroller4.GetUserStatsByYear)                            // chart per year
	app.Get("/display/users", admincontroller2.GetFilteredUserDetails)                                  // Fetch all users can be filtered through name=John,accountname=artem&user_type=Landlord                                //# Search by fullname GET /users/search?field=fullname&search_term=Artem# Search by email	GET /users/search?field=email&search_term=example.com # Search by phone number GET /users/search?field=phone_number&search_term=+12345
	app.Get("/admin/count/:user_type", admincontroller2.CountUsersByType)                               //displaying number of users by usertype
	app.Get("/admin/count-user/:account_status/:user_type", admincontroller2.CountUsersByStatusAndType) //displaying number of users whose verified and still pending
	app.Get("/admin/count_apartment/:status", admincontroller2.CountApartmentsByStatus)                 //displaying number of users by usertype
	app.Get("/admin/count-property-type/:property_type", admincontroller2.CountApartmentsByPropertyType)
	app.Get("/admin/count-apartment/:status/:property_type", admincontroller2.CountApartmentsByStatusAndType) //displaying toal number of both pending & property type
	app.Get("/admin/apartments/details", admincontroller3.GetFilteredApartments)                              //Get complete apartment details along with other data and can be filtered
	app.Get("/apartments/pending", admincontroller.GetPendingApartments)                                      // Fetch unverified apartments
	app.Get("/user/pending", admincontroller.GetPendingUsers)                                                 // Fetch unverified users
	app.Get("/admin/apartmentfilter", admincontroller2.Apartmentfilteradmin)
	app.Get("/landlord/profileid/:uid", admincontroller2.GetLatestLandlordID)

	//////////////////// DELETE //////////////////
	app.Delete("/admin/apartment/delete/:id", admincontroller3.DeleteApartmentByID) // Delete speific apartment
	app.Delete("/admin/user/:uid", admincontroller2.SoftDeleteUser)                 // Mark the account status as deleted

	//////////////////// Admin //////////////////

	//////////////////// FOR ALL //////////////////

	//////////////////// PUT //////////////////
	app.Put("/api/user/update-contact", middleware.AuthMiddleware, landlordcontroller2.UpdateContactInfo)

	//////////////////// POST //////////////////
	app.Post("/create/validid", middleware.AuthMiddleware, all.SetValidID)
	app.Post("/create/validid", middleware.AuthMiddleware, all.SetValidID)
	app.Post("/signup", authcontroller.Signup) // Register a new us
	//////////////////// GET //////////////////
	app.Get("/all/filter-apartments/", all.FetchApprovedApartmentsForTenant) //http://localhost:3000/all/filter-apartments?amenities=Wifi,Laundry&house_rules=No Smoking&min_price=3000&max_price=8000&property_types=Condo,Apartment
	app.Get("/allapartments/search", all.SearchApartments)
	app.Get("/all/apartmentfulldetails/:id", all.FetchSingleApartmentDetails) // view all of the specific apartment details

	//////////////////// FOR ALL //////////////////

	//////////////////// Tenant //////////////////

	//////////////////// PUT //////////////////

	//////////////////// POST //////////////////
	app.Post("/create/inquiry", middleware.AuthMiddleware, tenantscontroller.CreateInquiry)
	// app.Post("/tenant/delete-inquiry", middleware.AuthMiddleware, tenantscontroller.DeleteInquiryAfterViewingNotification)
	app.Post("/add/wishlist", middleware.AuthMiddleware, tenantscontroller.AddToWishlist)
	app.Post("/add/recentlyviewed", middleware.AuthMiddleware, tenantscontroller.AddToRecentlyViewed)
	app.Get("/get/recently-viewed", middleware.AuthMiddleware, controller.FetchRecentlyViewed)
	//////////////////// GET //////////////////
	// app.Get("/tenant/inquiries/count-status", middleware.AuthMiddleware, tenantscontroller.CountAcceptedOrRejectedInquiries)
	// app.Get("/tenant/inquiries/get-notification", middleware.AuthMiddleware, tenantscontroller.GetAllinquiries) // Display all inquiries
	app.Get("/api/apartments/Approved", tenantscontroller.FetchApprovedApartmentsForTenant) //Display all the Approved apartment
	app.Get("/get/wishlist", middleware.AuthMiddleware, tenantscontroller.FetchwishlistForTenant)

	//////////////////// DELETE //////////////////
	app.Delete("/wishlist/:apartment_id", middleware.AuthMiddleware, tenantscontroller.RemoveFromWishlist)

	//////////////////// Tenant //////////////////

	app.Post("/firebase", authcontroller.VerifyFirebaseToken)

	// app.Post("/api/send-notification", func(c *fiber.Ctx) error {
	// 	type RequestBody struct {
	// 		FcmToken       string `json:"fcmToken"`
	// 		Title          string `json:"title"` // Optional, may be set dynamically
	// 		Body           string `json:"body"`
	// 		ConversationId string `json:"conversationId"`
	// 		SenderId       string `json:"senderId"`
	// 		Debug          bool   `json:"debug"`
	// 	}

	// 	var req RequestBody
	// 	if err := c.BodyParser(&req); err != nil {
	// 		log.Printf("[Notification] Invalid request: %v", err)
	// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 			"error":   "Invalid request body",
	// 			"details": err.Error(),
	// 		})
	// 	}

	// 	// Validate required fields
	// 	if req.FcmToken == "" || req.ConversationId == "" {
	// 		log.Printf("[Notification] Missing fields: Token provided? %t, Conversation ID provided? %t",
	// 			req.FcmToken != "", req.ConversationId != "")
	// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 			"error": "Missing required fields (fcmToken and conversationId are required)",
	// 		})
	// 	}

	// 	// 🧠 Dynamically set title if not provided

	// 	log.Printf("[Notification] Sending to token: %s | Conversation ID: %s",
	// 		maskToken(req.FcmToken), req.ConversationId)

	// 	// Call the notification sending logic
	// 	config.SendPushNotification(
	// 		req.FcmToken,
	// 		req.Title,
	// 		req.Body,
	// 		req.ConversationId,
	// 		req.SenderId,
	// 	)

	// 	// Success response
	// 	response := fiber.Map{
	// 		"status":  "success",
	// 		"message": "Notification sent successfully",
	// 		"data": fiber.Map{
	// 			"conversationId": req.ConversationId,
	// 			"timestamp":      time.Now().Format(time.RFC3339),
	// 		},
	// 	}

	// 	// Optional debug response
	// 	if req.Debug {
	// 		response["debug"] = fiber.Map{
	// 			"fcmToken": maskToken(req.FcmToken),
	// 			"senderId": req.SenderId,
	// 		}
	// 	}

	// 	log.Printf("[Notification] Sent successfully | Conversation ID: %s", req.ConversationId)
	// 	return c.JSON(response)
	// })

	app.Post("/api/send-notification", func(c *fiber.Ctx) error {
		type RequestBody struct {
			FcmToken       string `json:"fcmToken"`
			Title          string `json:"title"` // Optional, may be set dynamically
			Body           string `json:"body"`
			ConversationId string `json:"conversationId"` // Optional now
			SenderId       string `json:"senderId"`
			Debug          bool   `json:"debug"`
		}

		var req RequestBody
		if err := c.BodyParser(&req); err != nil {
			log.Printf("[Notification] Invalid request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
		}

		// Validate required fields
		if req.FcmToken == "" {
			log.Println("[Notification] Missing required field: fcmToken")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required field: fcmToken",
			})
		}

		// Optional: Generate a default conversation ID if needed
		if req.ConversationId == "" {
			req.ConversationId = "general" // Or generate a UUID if needed
		}

		// Dynamically set title if not provided
		if req.Title == "" {
			req.Title = "New Notification"
		}

		log.Printf("[Notification] Sending to token: %s | Conversation ID: %s",
			maskToken(req.FcmToken), req.ConversationId)

		// Call the notification sending logic
		config.SendPushNotification(
			req.FcmToken,
			req.Title,
			req.Body,
			req.ConversationId,
			req.SenderId,
		)

		// Success response
		response := fiber.Map{
			"status":  "success",
			"message": "Notification sent successfully",
			"data": fiber.Map{
				"conversationId": req.ConversationId,
				"timestamp":      time.Now().Format(time.RFC3339),
			},
		}

		// Optional debug response
		if req.Debug {
			response["debug"] = fiber.Map{
				"fcmToken": maskToken(req.FcmToken),
				"senderId": req.SenderId,
			}
		}

		log.Printf("[Notification] Sent successfully | Conversation ID: %s", req.ConversationId)
		return c.JSON(response)
	})

	app.Post("/api/track-open/:logId", handlers.TrackNotificationOpenHandler)
	app.Get("/notifications/:uid", config.GetNotificationsHandler)

	app.Get("/unverifyAndResend/:uid", config.UnverifyAndResendHandler)

	ratingGroup := app.Group("/api/ratings")
	ratingGroup.Post("/confirm", middleware.AuthMiddleware, controller.ConfirmRental)
	ratingGroup.Post("/submit", middleware.AuthMiddleware, controller.SubmitRating)
	ratingGroup.Get("/apartment/:id", controller.GetApartmentRatings)
	ratingGroup.Get("/tenant/:id", middleware.AuthMiddleware, controller.GetTenantIDByRentalAgreementID)
	ratingGroup.Get("/check", middleware.AuthMiddleware, controller.CheckRatingEligibility)
	app.Get("/api/inquiries/has-inquiry", middleware.AuthMiddleware, controller.CheckHasInquiry)

	//Payment using Gcash routes

}

// Helper to mask sensitive token in logs
func maskToken(token string) string {
	if len(token) < 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
