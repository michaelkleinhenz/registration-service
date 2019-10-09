package signup

import (
	"fmt"
	"strings"

	crtapi "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
	"github.com/codeready-toolchain/registration-service/pkg/kubeclient"

	errors2 "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/rest"
)

// SignupServiceConfiguration represents the config used for the signup service.
type SignupServiceConfiguration interface {
	GetNamespace() string
}

// SignupService represents the signup service for controllers.
type SignupService interface {
	GetUserSignup(userID string) (*Signup, error)
	CreateUserSignup(username, userID string) (*crtapi.UserSignup, error)
}

// SignupServiceImpl represents the implementation of the signup service.
type SignupServiceImpl struct {
	Namespace   string
	UserSignups kubeclient.UserSignupInterface
}

// NewSignupService creates a service object for performing user signup-related activities
func NewSignupService(cfg SignupServiceConfiguration) (SignupService, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubeclient.NewCRTV1Alpha1Client(k8sConfig, cfg.GetNamespace())
	if err != nil {
		return nil, err
	}

	return &SignupServiceImpl{
		Namespace:   cfg.GetNamespace(),
		UserSignups: client.UserSignups(),
	}, nil
}

// CreateUserSignup creates a new UserSignup resource with the specified username and userID
func (c *SignupServiceImpl) CreateUserSignup(username, userID string) (*crtapi.UserSignup, error) {
	name, err := c.transformAndValidateUserName(username)
	if err != nil {
		return nil, err
	}

	userSignup := &crtapi.UserSignup{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
		Spec: crtapi.UserSignupSpec{
			UserID:        userID,
			TargetCluster: "",
			Approved:      false,
			Username:      username,
		},
	}

	created, err := s.UserSignups.Create(userSignup)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// GetUserSignup gets the UserSignup resource with the specified userID
// Returns nil, nil if the resource is not found
func (s *SignupServiceImpl) GetUserSignup(userID string) (*Signup, error) {
	// TODO
	/*
				us, err := c.UserSignups.Get(userID)
			    // TODO Check if signup exists. If yes then get the corresponding MUR and populate the status
		        // transform crt.UserSignup to signup.Signup
	*/
	return nil, nil
}

func (s *SignupServiceImpl) transformAndValidateUserName(username string) (string, error) {
	replaced := strings.ReplaceAll(strings.ReplaceAll(username, "@", "-at-"), ".", "-")

	errs := validation.IsQualifiedName(replaced)
	if len(errs) > 0 {
		return "", errors2.New(fmt.Sprintf("Transformed username [%s] is invalid", username))
	}

	iteration := 0
	transformed := replaced

	for {
		userSignup, err := s.UserSignups.Get(transformed)
		if err != nil {
			if !errors.IsNotFound(err) {
				return "", err
			}
		}

		if userSignup == nil {
			break
		}

		iteration++
		transformed = fmt.Sprintf("%s-%d", replaced, iteration)
	}

	return transformed, nil
}
