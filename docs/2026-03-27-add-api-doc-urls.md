# Add API Documentation URL Comments

## Overview
Add `// See:` doc URL comments to every public method (regular + Raw) across all 33 service files, linking to the Intercom API reference. Also remove the deprecated `RunAssignmentRules` method.

## Context
- 33 service files, ~120 endpoints, currently zero doc URL comments
- URL pattern: `https://developers.intercom.com/docs/references/rest-api/api.intercom.io/{tag}/{operationId}` (both lowercase, tag spaces → dashes)
- For endpoints without operationId (ExportReporting): path-encoded URLs like `export/paths/~1export~1reporting_data~1enqueue/post`
- Source of truth: OpenAPI 2.15 spec + local `raw/api.intercom.io.yaml`

## Development Approach
- No tests needed — comment-only changes (except RunAssignmentRules removal which needs test cleanup)
- Run `go vet ./...` and `go test ./...` after all edits

## Comment Format
```go
// Get retrieves a contact by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/contacts/showcontact
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
```

## Skip Rules
- Parse functions (`ParseXxxResult`) — no endpoint
- ListAll methods — iterator wrappers
- Model structs — no per-model URL
- Service type declarations — just type aliases

## Base URL
`https://developers.intercom.com/docs/references/rest-api/api.intercom.io`

## Implementation Steps

### Task 1: Remove deprecated RunAssignmentRules
- [x] Remove `RunAssignmentRules` and `RunAssignmentRulesRaw` from `conversations.go`
- [x] Remove `ParseConversationRunAssignmentRulesResult` from `conversations.go`
- [x] Remove any related request/response types if they exist solely for this method
- [x] Remove test code for RunAssignmentRules from `conversations_test.go`
- [x] Run `go vet ./...` and `go test ./...`

### Task 2: Add doc URLs to admins.go
- [x] Me/MeRaw → `admins/identifyadmin`
- [x] Get/GetRaw → `admins/retrieveadmin`
- [x] List/ListRaw → `admins/listadmins`
- [x] SetAway/SetAwayRaw → `admins/setawayadmin`
- [x] ListActivityLogs/ListActivityLogsRaw → `admins/listactivitylogs`

### Task 3: Add doc URLs to ai_content.go
- [x] ListContentImportSources/Raw → `ai-content/listcontentimportsources`
- [x] GetContentImportSource/Raw → `ai-content/getcontentimportsource`
- [x] CreateContentImportSource/Raw → `ai-content/createcontentimportsource`
- [x] UpdateContentImportSource/Raw → `ai-content/updatecontentimportsource`
- [x] DeleteContentImportSource/Raw → `ai-content/deletecontentimportsource`
- [x] ListExternalPages/Raw → `ai-content/listexternalpages`
- [x] GetExternalPage/Raw → `ai-content/getexternalpage`
- [x] CreateExternalPage/Raw → `ai-content/createexternalpage`
- [x] UpdateExternalPage/Raw → `ai-content/updateexternalpage`
- [x] DeleteExternalPage/Raw → `ai-content/deleteexternalpage`

### Task 4: Add doc URLs to articles.go
- [x] Get/GetRaw → `articles/retrievearticle`
- [x] List/ListRaw → `articles/listarticles`
- [x] Create/CreateRaw → `articles/createarticle`
- [x] Update/UpdateRaw → `articles/updatearticle`
- [x] Delete/DeleteRaw → `articles/deletearticle`
- [x] Search/SearchRaw → `articles/searcharticles`

### Task 5: Add doc URLs to away_status_reasons.go
- [x] List/ListRaw → `away-status-reasons/listawaystatusreasons`

### Task 6: Add doc URLs to brands.go
- [x] Get/GetRaw → `brands/retrievebrand`
- [x] List/ListRaw → `brands/listbrands`

### Task 7: Add doc URLs to calls.go
- [x] Get/GetRaw → `calls/showcall`
- [x] List/ListRaw → `calls/listcalls`
- [x] Search/SearchRaw → `calls/listcallswithtranscripts`
- [x] GetRecordingURL/GetRecordingURLRaw → `calls/showcallrecording`
- [x] GetTranscript/GetTranscriptRaw → `calls/showcalltranscript`

### Task 8: Add doc URLs to companies.go
- [ ] Get/GetRaw → `companies/retrieveacompanybyid`
- [ ] List/ListRaw → `companies/retrievecompany`
- [ ] Create/CreateRaw → `companies/createorupdatecompany`
- [ ] Update/UpdateRaw → `companies/updatecompany`
- [ ] Delete/DeleteRaw → `companies/deletecompany`
- [ ] Scroll/ScrollRaw → `companies/scrolloverallcompanies`
- [ ] ListContacts/ListContactsRaw → `companies/listattachedcontacts`
- [ ] ListSegments/ListSegmentsRaw → `companies/listattachedsegmentsforcompanies`
- [ ] ListNotes/ListNotesRaw → `companies/listcompanynotes`
- [ ] CompanyList/CompanyListRaw → `companies/listallcompanies`

### Task 9: Add doc URLs to contacts.go
- [ ] Get/GetRaw → `contacts/showcontact`
- [ ] List/ListRaw → `contacts/listcontacts`
- [ ] Create/CreateRaw → `contacts/createcontact`
- [ ] Update/UpdateRaw → `contacts/updatecontact`
- [ ] Delete/DeleteRaw → `contacts/deletecontact`
- [ ] Search/SearchRaw → `contacts/searchcontacts`
- [ ] Merge/MergeRaw → `contacts/mergecontact`
- [ ] Archive/ArchiveRaw → `contacts/archivecontact`
- [ ] Unarchive/UnarchiveRaw → `contacts/unarchivecontact`
- [ ] Block/BlockRaw → `contacts/blockcontact`
- [ ] FindByExternalID/FindByExternalIDRaw → `contacts/showcontactbyexternalid`
- [ ] ListCompanies/ListCompaniesRaw → `contacts/listcompaniesforacontact`
- [ ] AddCompany/AddCompanyRaw → `contacts/attachcontacttoacompany`
- [ ] RemoveCompany/RemoveCompanyRaw → `contacts/detachcontactfromacompany`
- [ ] ListNotes/ListNotesRaw → `contacts/listnotes`
- [ ] CreateNote/CreateNoteRaw → `contacts/createnote`
- [ ] ListSegments/ListSegmentsRaw → `contacts/listsegmentsforacontact`
- [ ] ListSubscriptions/ListSubscriptionsRaw → `subscription-types/listsubscriptionsforacontact`
- [ ] AddSubscription/AddSubscriptionRaw → `subscription-types/attachsubscriptiontypetocontact`
- [ ] RemoveSubscription/RemoveSubscriptionRaw → `subscription-types/detachsubscriptiontypetocontact`
- [ ] AddTag/AddTagRaw → `tags/attachtagtocontact`
- [ ] RemoveTag/RemoveTagRaw → `tags/detachtagfromcontact`

### Task 10: Add doc URLs to conversations.go
- [ ] Get/GetRaw → `conversations/retrieveconversation`
- [ ] List/ListRaw → `conversations/listconversations`
- [ ] Create/CreateRaw → `conversations/createconversation`
- [ ] Update/UpdateRaw → `conversations/updateconversation`
- [ ] Delete/DeleteRaw → `conversations/deleteconversation`
- [ ] Search/SearchRaw → `conversations/searchconversations`
- [ ] Reply/ReplyRaw → `conversations/replyconversation`
- [ ] Close/CloseRaw → `conversations/manageconversation`
- [ ] Open/OpenRaw → `conversations/manageconversation`
- [ ] Snooze/SnoozeRaw → `conversations/manageconversation`
- [ ] Assign/AssignRaw → `conversations/manageconversation`
- [ ] Convert/ConvertRaw → `conversations/convertconversationtoticket`
- [ ] Redact/RedactRaw → `conversations/redactconversation`
- [ ] AddCustomer/AddCustomerRaw → `conversations/attachcontacttoconversation`
- [ ] RemoveCustomer/RemoveCustomerRaw → `conversations/detachcontactfromconversation`
- [ ] AddTag/AddTagRaw → `conversations/attachtagtoconversation`
- [ ] RemoveTag/RemoveTagRaw → `conversations/detachtagfromconversation`

### Task 11: Add doc URLs to custom_channel_events.go
- [ ] NotifyNewConversation/Raw → `custom-channel-events/notifynewconversation`
- [ ] NotifyNewMessage/Raw → `custom-channel-events/notifynewmessage`
- [ ] NotifyQuickReply/Raw → `custom-channel-events/notifyquickreplyselected`
- [ ] NotifyAttributeCollected/Raw → `custom-channel-events/notifyattributecollected`

### Task 12: Add doc URLs to custom_objects.go
- [ ] Get/GetRaw → `custom-object-instances/getcustomobjectinstancesbyid`
- [ ] GetByExternalID/GetByExternalIDRaw → `custom-object-instances/getcustomobjectinstancesbyexternalid`
- [ ] CreateOrUpdate/CreateOrUpdateRaw → `custom-object-instances/createcustomobjectinstances`
- [ ] Delete/DeleteRaw → `custom-object-instances/deletecustomobjectinstancesbyexternalid`
- [ ] DeleteByExternalID/DeleteByExternalIDRaw → `custom-object-instances/deletecustomobjectinstancesbyid`

### Task 13: Add doc URLs to data_attributes.go
- [ ] List/ListRaw → `data-attributes/lisdataattributes`
- [ ] Create/CreateRaw → `data-attributes/createdataattribute`
- [ ] Update/UpdateRaw → `data-attributes/updatedataattribute`

### Task 14: Add doc URLs to data_events.go
- [ ] Create/CreateRaw → `data-events/createdataevent`
- [ ] List/ListRaw → `data-events/lisdataevents`
- [ ] CreateSummaries/CreateSummariesRaw → `data-events/dataeventsummaries`

### Task 15: Add doc URLs to data_export.go
- [ ] Create/CreateRaw → `data-export/createdataexport`
- [ ] GetStatus/GetStatusRaw → `data-export/getdataexport`
- [ ] Cancel/CancelRaw → `data-export/canceldataexport`

### Task 16: Add doc URLs to emails.go
- [ ] Get/GetRaw → `emails/retrieveemail`
- [ ] List/ListRaw → `emails/listemails`

### Task 17: Add doc URLs to export_reporting.go (path-encoded URLs)
- [ ] Enqueue/EnqueueRaw → `export/paths/~1export~1reporting_data~1enqueue/post`
- [ ] GetStatus/GetStatusRaw → `export/paths/~1export~1reporting_data~1{job_identifier}/get`
- [ ] GetDatasets/GetDatasetsRaw → `export/paths/~1export~1reporting_data~1get_datasets/get`
- [ ] Download → `export/paths/~1download~1reporting_data~1{job_identifier}/get`

### Task 18: Add doc URLs to fin_voice.go
- [ ] Register/RegisterRaw → `calls/registerfinvoicecall`
- [ ] Collect/CollectRaw → `calls/collectfinvoicecallbyid`
- [ ] GetByExternalID/GetByExternalIDRaw → `calls/collectfinvoicecallbyexternalid`
- [ ] GetByConversationID/GetByConversationIDRaw → `calls/collectfinvoicecallsbyconversationid`
- [ ] GetByPhoneNumber/GetByPhoneNumberRaw → `calls/collectfinvoicecallbyphonenumber`

### Task 19: Add doc URLs to help_center.go
- [ ] ListCollections/ListCollectionsRaw → `help-center/listallcollections`
- [ ] GetCollection/GetCollectionRaw → `help-center/retrievecollection`
- [ ] CreateCollection/CreateCollectionRaw → `help-center/createcollection`
- [ ] UpdateCollection/UpdateCollectionRaw → `help-center/updatecollection`
- [ ] DeleteCollection/DeleteCollectionRaw → `help-center/deletecollection`
- [ ] ListHelpCenters/ListHelpCentersRaw → `help-center/listhelpcenters`
- [ ] GetHelpCenter/GetHelpCenterRaw → `help-center/retrievehelpcenter`

### Task 20: Add doc URLs to internal_articles.go
- [ ] Get/GetRaw → `internal-articles/retrieveinternalarticle`
- [ ] List/ListRaw → `internal-articles/listinternalarticles`
- [ ] Create/CreateRaw → `internal-articles/createinternalarticle`
- [ ] Update/UpdateRaw → `internal-articles/updateinternalarticle`
- [ ] Delete/DeleteRaw → `internal-articles/deleteinternalarticle`
- [ ] Search/SearchRaw → `internal-articles/searchinternalarticles`

### Task 21: Add doc URLs to ip_allowlist.go
- [ ] Get/GetRaw → `ip-allowlist/getipallowlist`
- [ ] Update/UpdateRaw → `ip-allowlist/updateipallowlist`

### Task 22: Add doc URLs to jobs.go
- [ ] GetStatus/GetStatusRaw → `jobs/jobsstatus`

### Task 23: Add doc URLs to messages.go
- [ ] Create/CreateRaw → `messages/createmessage`

### Task 24: Add doc URLs to news.go
- [ ] ListNewsItems/ListNewsItemsRaw → `news/listnewsitems`
- [ ] GetNewsItem/GetNewsItemRaw → `news/retrievenewsitem`
- [ ] CreateNewsItem/CreateNewsItemRaw → `news/createnewsitem`
- [ ] UpdateNewsItem/UpdateNewsItemRaw → `news/updatenewsitem`
- [ ] DeleteNewsItem/DeleteNewsItemRaw → `news/deletenewsitem`
- [ ] ListNewsfeeds/ListNewsfeedsRaw → `news/listnewsfeeds`

### Task 25: Add doc URLs to notes.go
- [ ] Get/GetRaw → `notes/retrievenote`

### Task 26: Add doc URLs to phone_call_redirects.go
- [ ] Create/CreateRaw → `switch/createphoneswitch`

### Task 27: Add doc URLs to segments.go
- [ ] Get/GetRaw → `segments/retrievesegment`
- [ ] List/ListRaw → `segments/listsegments`

### Task 28: Add doc URLs to subscription_types.go
- [ ] List/ListRaw → `subscription-types/listsubscriptiontypes`

### Task 29: Add doc URLs to tags.go
- [ ] Get/GetRaw → `tags/findtag`
- [ ] List/ListRaw → `tags/listtags`
- [ ] CreateOrUpdate/CreateOrUpdateRaw → `tags/createtag`
- [ ] Delete/DeleteRaw → `tags/deletetag`
- [ ] TagCompany/TagCompanyRaw → `tags/createtag`
- [ ] UntagCompany/UntagCompanyRaw → `tags/createtag`

### Task 30: Add doc URLs to teams.go
- [ ] Get/GetRaw → `teams/retrieveteam`
- [ ] List/ListRaw → `teams/listteams`

### Task 31: Add doc URLs to ticket_states.go
- [ ] List/ListRaw → `ticket-states/listticketstates`

### Task 32: Add doc URLs to ticket_types.go
- [ ] Get/GetRaw → `ticket-types/gettickettype`
- [ ] List/ListRaw → `ticket-types/listtickettypes`
- [ ] Create/CreateRaw → `ticket-types/createtickettype`
- [ ] Update/UpdateRaw → `ticket-types/updatetickettype`
- [ ] ListAttributes/ListAttributesRaw → `ticket-type-attributes/createtickettypeattribute`
- [ ] CreateAttribute/CreateAttributeRaw → `ticket-type-attributes/createtickettypeattribute`
- [ ] UpdateAttribute/UpdateAttributeRaw → `ticket-type-attributes/updatetickettypeattribute`

### Task 33: Add doc URLs to tickets.go
- [ ] Get/GetRaw → `tickets/getticket`
- [ ] Create/CreateRaw → `tickets/createticket`
- [ ] Update/UpdateRaw → `tickets/updateticket`
- [ ] Delete/DeleteRaw → `tickets/deleteticket`
- [ ] Search/SearchRaw → `tickets/searchtickets`
- [ ] Reply/ReplyRaw → `tickets/replyticket`
- [ ] AddTag/AddTagRaw → `tags/attachtagtoticket`
- [ ] RemoveTag/RemoveTagRaw → `tags/detachtagfromticket`
- [ ] Enqueue/EnqueueRaw → `tickets/enqueuecreateticket`

### Task 34: Add doc URLs to visitors.go
- [ ] Get/GetRaw → `visitors/retrievevisitorwithuserid`
- [ ] Update/UpdateRaw → `visitors/updatevisitor`
- [ ] FindByUserID/FindByUserIDRaw → `visitors/retrievevisitorwithuserid`
- [ ] Delete/DeleteRaw → `visitors/convertvisitor`

### Task 35: Final verification
- [ ] Run `go vet ./...`
- [ ] Run `go test ./...`
- [ ] Count added comments: `grep -r "// See: https://developers.intercom.com" *.go | wc -l`
- [ ] Spot-check a few URLs manually

## Post-Completion
- No deployment needed — comment-only changes (plus deprecated method removal)
