GOOGLE_CLOUD_PROJECT=""
GOOGLE_REGION="asia-east1"
CLOUD_RUN_SERVICE=""
SECRET_PROJECT_ID=""
SECRET_NAME=""
LINE_API_ENDPOINT="https://api.line.me/v2/bot/message/push"

docker build -t "$GOOGLE_REGION-docker.pkg.dev/$GOOGLE_CLOUD_PROJECT/api/$CLOUD_RUN_SERVICE" .

docker push "$GOOGLE_REGION-docker.pkg.dev/$GOOGLE_CLOUD_PROJECT/api/$CLOUD_RUN_SERVICE"

gcloud run deploy $CLOUD_RUN_SERVICE --image "$GOOGLE_REGION-docker.pkg.dev/$GOOGLE_CLOUD_PROJECT/api/$CLOUD_RUN_SERVICE" \
    --platform managed \
    --region $GOOGLE_REGION \
    --update-env-vars LINE_API_ENDPOINT=$LINE_API_ENDPOINT,SECRET_PROJECT_ID=$SECRET_PROJECT_ID,SECRET_NAME=$SECRET_NAME \
    --allow-unauthenticated