package mock

//go:generate minimock -g -i github.com/weareyolo/cloudmetrics.CloudWatch -o ./ -s "_mock.go"
//go:generate minimock -g -i github.com/weareyolo/cloudmetrics.DatumBuilder -o ./ -s "_mock.go"
//go:generate minimock -g -i github.com/weareyolo/cloudmetrics.Publisher -o ./ -s "_mock.go"
