package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	rootCmd.AddCommand(cmdCreateStaff)
}

func prompt(reader *bufio.Reader, message string, required bool) string {
	var p string
	for {
		fmt.Print(message)
		input, _ := reader.ReadString('\n')
		p = strings.TrimSpace(input)
		if !required || p != "" {
			break
		}
	}
	return p
}

func isValidDateFormat(input string) bool {
	_, err := time.Parse(constants.DateLayoutISO, input)
	return err == nil
}

func isValidTimeFormat(input string) bool {
	_, err := time.Parse(constants.TimeLayoutHHMM, input)
	return err == nil
}

var cmdCreateStaff = &cobra.Command{
	Use:   "create_staff",
	Short: "Create a new staff member",
	Long:  "Create a new staff member with interactive prompts",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		firstName := prompt(reader, "First Name: ", true)
		middleName := prompt(reader, "Middle Name (optional): ", false)
		lastName := prompt(reader, "Last Name: ", true)
		birthdate := prompt(reader, "Birthdate (YYYY-MM-DD): ", true)
		sex := strings.ToLower(prompt(reader, "Sex (M/F): ", true))
		if sex != "m" && sex != "f" {
			panic("sex must be either M or F")
		}
		dateHired := prompt(reader, "Date Hired (YYYY-MM-DD): ", true)
		if !isValidDateFormat(dateHired) {
			panic("not valid YYYY-MM-DD date")
		}
		position := prompt(reader, "Position: ", true)
		userTypeInput := prompt(reader, "User Type (staff/superuser): ", true)
		userType := enums.MustParseStaffUserTypeToEnum(strings.ToUpper(userTypeInput))
		email := prompt(reader, "Email: ", true)
		if !constants.EmailRegex.MatchString(email) {
			panic("must be a valid email")
		}

		mobileNo := prompt(reader, "Mobile No: ", true)
		timeInSched := prompt(reader, "Time in schedule (HH:MM): ", true)
		if !isValidTimeFormat(timeInSched) {
			panic("not valid HH:MM time")
		}
		timeOutSched := prompt(reader, "Time out schedule (HH:MM): ", true)
		if !isValidTimeFormat(timeOutSched) {
			panic("not valid HH:MM time")
		}

		password := prompt(reader, "Password: ", true)
		if !constants.PasswordRegex.MatchString(password) {
			panic("must be alphanumeric [a-z A-Z 0-9 - _ . ? # @]")
		}

		requireInShopInput := prompt(reader, "Require in shop for time in/out (Y/N): ", true)
		requireInShop := strings.ToUpper(requireInShopInput) == "Y"

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to hash password: %v\n", err)
			os.Exit(1)
		}

		db := database.New(database.DB_MODE_RW)
		defer db.Close()

		middleNameNull := sql.NullString{String: middleName, Valid: middleName != ""}
		_, err = db.GetQueries().CreateStaff(cmd.Context(), queries.CreateStaffParams{
			FirstName:       firstName,
			MiddleName:      middleNameNull,
			LastName:        lastName,
			Birthdate:       birthdate,
			Sex:             strings.ToUpper(sex),
			DateHired:       dateHired,
			TimeInSchedule:  sql.NullString{Valid: true, String: timeInSched},
			TimeOutSchedule: sql.NullString{Valid: true, String: timeOutSched},
			Position:        position,
			UserType:        userType.String(),
			Email:           email,
			MobileNo:        mobileNo,
			Password:        string(hash),
			RequireInShop:   requireInShop,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create staff: %v\n", err)
			os.Exit(1)
		}

		logs.Log().Info(
			"Staff created",
			zap.String("first name", firstName),
			zap.String("middle name", middleName),
			zap.String("last name", lastName),
			zap.String("birthdate", birthdate),
			zap.String("sex", sex),
			zap.String("date hired", dateHired),
			zap.String("time in schedule", timeInSched),
			zap.String("time out schedule", timeOutSched),
			zap.String("position", position),
			zap.Stringer("usertype", userType),
			zap.String("email", email),
			zap.String("mobile number", mobileNo),
			zap.Bool("require in shop", requireInShop),
		)
	},
}
