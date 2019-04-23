package load

import (
	//"fmt"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

type Customer struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string
}

type User struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Email      string        `bson:"email,omitempty"`
	Username   string        `bson:"username,omitempty"`
	Admin      int           `bson:"admin,omitempty"`
	CustomerId bson.ObjectId `bson:"customerId,omitempty"`
	Password   string        `bson:"password,omitempty"`
}

type Status struct {
	Draft int
	App   int
}

type Goal struct {
	ID                  bson.ObjectId   `bson:"_id,omitempty"`
	OwnerId             bson.ObjectId   `bson:"ownerId,omitempty"`
	CreatorId           bson.ObjectId   `bson:"creatorId,omitempty"`
	CustomerId          bson.ObjectId   `bson:"customerId,omitempty"`
	ApproverId          bson.ObjectId   `bson:"approverId,omitempty"`
	ContributesTo       []bson.ObjectId `bson:"contributesTo,omitempty"`
	DependsOn           []bson.ObjectId `bson:"dependsOn,omitempty"`
	Title               string          `bson:"title,omitempty"`
	Description         string          `bson:"description,omitempty"`
	KeyResult           string          `bson:"keyResult,omitempty"`
	KeyResultScoreOne   string          `bson:"keyResultScoreOne,omitempty"`
	KeyResultScoreTwo   string          `bson:"keyResultScoreTwo,omitempty"`
	KeyResultScoreThree string          `bson:"keyResultScoreThree,omitempty"`
	KeyResultScoreFour  string          `bson:"keyResultScoreFour,omitempty"`
	KeyResultScoreFive  string          `bson:"keyResultScoreFive,omitempty"`
	StartDate           time.Time       `bson:"startDate,omitempty"`
	EndDate             time.Time       `bson:"endDate,omitempty"`
	Status              Status          `bson:"status,omitempty"`
}

func insertGoal(Goals mgo.Collection, user User, customer Customer, j int, done chan bool) {
	// add new Goals for that User
	//log.Print("Inserting a new Goal document")
	err := Goals.Insert(Goal{OwnerId: user.ID, CustomerId: customer.ID, CreatorId: user.ID, ApproverId: user.ID, Status: Status{1, 0}, Title: fmt.Sprintf("Goal %v", j)})
	if err != nil {
		log.Fatal(err)
	}
	//log.Print("Success: Goal document inserted")

	// notify that the insertGoal job is done
	if j == 999 {
		done <- true
	}
}

func Data() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// save Collection references
	Customers := session.DB("load-aw-mvp").C("Customers")
	Users := session.DB("load-aw-mvp").C("Users")
	Goals := session.DB("load-aw-mvp").C("Goals")

	// add a new Customer
	log.Print("Inserting a new Customer document")
	err = Customers.Insert(Customer{Name: "loadTestCustomer"})
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Success: Customer document inserted")
	log.Print("Finding new Customer ID")
	customer := Customer{}
	err = Customers.Find(Customer{Name: "loadTestCustomer"}).One(&customer)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Success: Customer ID = %v", customer.ID)

	// add new Users
	for i := 0; i < 100; i++ {
		insertGoalDone := make(chan bool, 1)
		//log.Print("Inserting a new User document")
		password, err := bcrypt.GenerateFromPassword([]byte("testpassword"), 10)
		err = Users.Insert(User{Email: fmt.Sprintf("%v@align.work", i), Username: "loadtest", Admin: 1, CustomerId: customer.ID, Password: string(password)})
		if err != nil {
			log.Fatal(err)
		}
		user := User{}
		err = Users.Find(User{Email: fmt.Sprintf("%v@align.work", i)}).One(&user)
		if err != nil {
			log.Fatal(err)
		}
		//log.Printf("Success: User ID = %v", user.ID)
		for j := 0; j < 1000; j++ {
			go insertGoal(*Goals, user, customer, j, insertGoalDone)
		}
		<-insertGoalDone
	}

	count, err := Goals.Count()
	log.Print("Routine complete - " + fmt.Sprintf("%v Goals created.", count))
	err = session.DB("load-aw-mvp").DropDatabase()

	// comment out this line if you want to keep the db
	log.Print("DB dropped.")
}
