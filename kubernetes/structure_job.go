package kubernetes

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	batchv1 "k8s.io/api/batch/v1"
)

func flattenJobSpec(in batchv1.JobSpec, d *schema.ResourceData, prefix ...string) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.ActiveDeadlineSeconds != nil {
		att["active_deadline_seconds"] = *in.ActiveDeadlineSeconds
	}

	if in.BackoffLimit != nil {
		att["backoff_limit"] = *in.BackoffLimit
	}

	if in.Completions != nil {
		att["completions"] = *in.Completions
	}

	if in.CompletionMode != nil {
		att["completion_mode"] = string(*in.CompletionMode)
	}

	if in.ManualSelector != nil {
		att["manual_selector"] = *in.ManualSelector
	}

	if in.Parallelism != nil {
		att["parallelism"] = *in.Parallelism
	}

	if in.Selector != nil {
		att["selector"] = flattenLabelSelector(in.Selector)
	}
	// Remove server-generated labels
	labels := in.Template.ObjectMeta.Labels

	if _, ok := labels["controller-uid"]; ok {
		delete(labels, "controller-uid")
	}

	if _, ok := labels["job-name"]; ok {
		delete(labels, "job-name")
	}

	podSpec, err := flattenPodTemplateSpec(in.Template, d, prefix...)
	if err != nil {
		return nil, err
	}
	att["template"] = podSpec

	if in.TTLSecondsAfterFinished != nil {
		att["ttl_seconds_after_finished"] = strconv.Itoa(int(*in.TTLSecondsAfterFinished))
	}

	return []interface{}{att}, nil
}

func expandJobSpec(j []interface{}) (batchv1.JobSpec, error) {
	obj := batchv1.JobSpec{}

	if len(j) == 0 || j[0] == nil {
		return obj, nil
	}

	in := j[0].(map[string]interface{})

	if v, ok := in["active_deadline_seconds"].(int); ok && v > 0 {
		obj.ActiveDeadlineSeconds = ptrToInt64(int64(v))
	}

	if v, ok := in["backoff_limit"].(int); ok && v != 6 {
		obj.BackoffLimit = ptrToInt32(int32(v))
	}

	if v, ok := in["completions"].(int); ok && v > 0 {
		obj.Completions = ptrToInt32(int32(v))
	}

	if v, ok := in["completion_mode"].(string); ok && v != "" {
		m := batchv1.CompletionMode(v)
		obj.CompletionMode = &m
	}

	if v, ok := in["manual_selector"]; ok {
		obj.ManualSelector = ptrToBool(v.(bool))
	}

	if v, ok := in["parallelism"].(int); ok && v >= 0 {
		obj.Parallelism = ptrToInt32(int32(v))
	}

	if v, ok := in["selector"].([]interface{}); ok && len(v) > 0 {
		obj.Selector = expandLabelSelector(v)
	}

	template, err := expandPodTemplate(in["template"].([]interface{}))
	if err != nil {
		return obj, err
	}
	obj.Template = *template

	if v, ok := in["ttl_seconds_after_finished"].(string); ok && v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return obj, err
		}
		obj.TTLSecondsAfterFinished = ptrToInt32(int32(i))
	}

	return obj, nil
}

func patchJobSpec(pathPrefix, prefix string, d *schema.ResourceData) (PatchOperations, error) {
	ops := make([]PatchOperation, 0)

	if d.HasChange(prefix + "active_deadline_seconds") {
		v := d.Get(prefix + "active_deadline_seconds").(int)
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/activeDeadlineSeconds",
			Value: v,
		})
	}

	if d.HasChange(prefix + "backoff_limit") {
		v := d.Get(prefix + "backoff_limit").(int)
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/backoffLimit",
			Value: v,
		})
	}

	if d.HasChange(prefix + "manual_selector") {
		v := d.Get(prefix + "manual_selector").(bool)
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/manualSelector",
			Value: v,
		})
	}

	if d.HasChange(prefix + "parallelism") {
		v := d.Get(prefix + "parallelism").(int)
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/parallelism",
			Value: v,
		})
	}

	return ops, nil
}
